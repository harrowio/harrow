package mailDispatcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	stdlibmail "net/mail"
	"net/textproto"
	"os"
	"strconv"
	"time"

	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/hmail"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/rs/zerolog"

	"github.com/jmoiron/sqlx"
	"github.com/streadway/amqp"
)

const (
	outputQueueName = "work.mail.outbound.all"
)

type ContextUser struct {
	Uuid    string `json:"uuid"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	UrlHost string `json:"url_host"`
}

var (
	ErrMalformedPayload   = errors.New("Malformed payload.")
	ErrMarshallingPayload = errors.New("Failed to marshal payload.")
	ErrStorage            = errors.New("Failed to query data.")
	ErrInternalOperation  = errors.New("Internal operation (e.g. repo check) can't email about this.")
)

func handleWork(log logger.Logger, fromAddress string, c *config.Config, db *sqlx.DB, out *amqp.Channel, message broadcast.Message) error {
	log.Info().Msgf("Receive: %s", message)
	if message.Table() != "activities" {
		message.RejectForever()
		return nil
	}

	tx := db.MustBegin()
	defer tx.Commit()

	activityStore := stores.NewDbActivityStore(tx)
	activityId, err := strconv.Atoi(message.UUID())
	if err != nil {
		log.Error().Msgf(" strconv.Atoi(message.UUID()): %s", err)
		return message.RejectForever()
	}
	activity, err := activityStore.FindActivityById(activityId)
	if err != nil {
		log.Info().Msgf("activityStore.FindActivityById(%d): %s", activityId, err)
		return message.RejectForever()
	}

	handler, found := activityHandlers[activity.Name]
	if !found {
		return message.RejectForever()
	}

	log.Info().Msgf("Handle: activity %q", activity.Name)

	mails, err := handler(log, c, activity, tx)
	if err != nil {
		if err == ErrMalformedPayload {
			log.Info().Msgf("Drop: %s", err)
			return message.RejectForever()
		}

		log.Info().Msgf("Error: %s", err)
		if err == ErrInternalOperation {
			log.Info().Msgf("Internal Operation on the bus, nacking permanently")
			return message.RejectForever()
		} else {
			return message.RequeueAfter(10 * time.Second)
		}
	}

	if err := message.Acknowledge(); err != nil {
		return err
	}

	if mails == nil {
		return nil
	}

	for _, mail := range mails {
		log.Info().Msgf("sendMail: %s -> %#v", message, mail.To)

		err := sendMail(fromAddress, out, mail)
		switch e := err.(type) {
		case *net.OpError, *amqp.Error:
			return e
		case nil:
			// noop
		default:
			log.Info().Msgf("sendMail: %s, equeuing in 10s", e)
			return message.RequeueAfter(10 * time.Second)
		}
	}

	return nil
}

func sendMail(fromAddress string, out *amqp.Channel, mail *hmail.Mail) error {
	if mail.Headers == nil {
		mail.Headers = textproto.MIMEHeader{}
	}

	mail.Headers.Set("Subject", mail.Data.Recipient.Subject)
	from := stdlibmail.Address{"Harrow.io", fromAddress}
	mail.From = from.String()

	body, err := json.Marshal(mail)
	if err != nil {
		log.Info().Msgf("sendMail: %s", err)
		return ErrMarshallingPayload
	}

	log.Info().Msgf("Recipient: %#v\n", mail.Data.Recipient)
	log.Info().Msgf("Mail: %#v\n", mail)
	log.Info().Msgf("Sending Message: %s", body)

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
	}

	if err := out.Publish(outputQueueName, mail.RoutingKey, true, false, msg); err != nil {
		log.Error().Msgf("sendMail: error publishing message: %s\n%s", err, body)

		return err
	} else {
		log.Info().Msgf("sendMail: published to %s:\n%s", outputQueueName, body)

		return nil
	}
}

type PgTime time.Time

func (t *PgTime) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		data = data[1 : len(data)-1]
	}

	parsed, err := time.Parse(time.RFC3339Nano, string(data))
	if err != nil {
		parsed, err = time.Parse("2006-01-02 15:04:05.999999999-07", string(data))
	}

	if err != nil {
		log.Info().Msgf("PgTime: UnmarshalJSON(%s): %s", string(data), err)
		return err
	}

	*t = PgTime(parsed)

	return nil
}

type OperationData struct {
	Uuid       string  `json:"uuid"`
	ExitStatus int     `json:"exit_status"`
	FinishedAt *PgTime `json:"finished_at"`
	TimedOutAt *PgTime `json:"timed_out_at"`
	StartedAt  *PgTime `json:"started_at"`
}

type OperationChangedData struct {
	Table string         `json:"table"`
	New   *OperationData `json:"new"`
	Old   *OperationData `json:"old"`

	operation  *domain.Operation
	project    *domain.Project
	job        *domain.Job
	recipients []*domain.User
}

func handleOperationChangedToStarted(data *OperationChangedData) ([]*hmail.Mail, error) {
	mails := []*hmail.Mail{}

	actor := &hmail.Actor{
		DisplayName: "An operation",
	}

	action := &hmail.Action{
		DisplayName: "started",
	}

	object := &hmail.Object{
		DisplayName: fmt.Sprintf("%s/%s", data.project.Name, data.job.Name),
		Uri:         fmt.Sprintf("/a/operations/%s", data.operation.Uuid),
	}

	subject := fmt.Sprintf("%s/%s started on %s",
		data.project.Name, data.job.Name,
		(func() string {
			if data.operation.StartedAt == nil {
				return "?"
			}
			return data.operation.StartedAt.Format(time.Stamp)
		})(),
	)

	for _, user := range data.recipients {
		recipient := &hmail.Recipient{
			DisplayName: user.Name,
			Subject:     subject,
			UrlHost:     user.UrlHost,
		}

		mails = append(mails, &hmail.Mail{
			RoutingKey: domain.EventOperationStarted,
			To:         []string{user.Email},
			Headers: textproto.MIMEHeader{
				"Subject": []string{subject},
			},
			Data: &hmail.MailContext{
				Actor:     actor,
				Action:    action,
				Object:    object,
				Recipient: recipient,
			},
		})
	}

	return mails, nil
}

func handleOperationChangedToTimedOut(data *OperationChangedData) ([]*hmail.Mail, error) {
	mails := []*hmail.Mail{}

	actor := &hmail.Actor{
		DisplayName: "An operation",
	}

	action := &hmail.Action{
		DisplayName: "timed out",
	}

	object := &hmail.Object{
		DisplayName: fmt.Sprintf("%s/%s", data.project.Name, data.job.Name),
		Uri:         fmt.Sprintf("/a/operations/%s", data.operation.Uuid),
	}

	subject := fmt.Sprintf("%s/%s timed out on %s",
		data.project.Name, data.job.Name,
		data.operation.TimedOutAt.Format(time.Stamp),
	)

	for _, user := range data.recipients {
		recipient := &hmail.Recipient{
			DisplayName: user.Name,
			Subject:     subject,
			UrlHost:     user.UrlHost,
		}

		mails = append(mails, &hmail.Mail{
			RoutingKey: domain.EventOperationTimedOut,
			To:         []string{user.Email},
			Headers: textproto.MIMEHeader{
				"Subject": []string{subject},
			},
			Data: &hmail.MailContext{
				Actor:     actor,
				Action:    action,
				Object:    object,
				Recipient: recipient,
			},
		})
	}

	return mails, nil
}

func handleOperationChangedToFinished(data *OperationChangedData) ([]*hmail.Mail, error) {
	mails := []*hmail.Mail{}

	actionStr := "failed"
	routingKey := domain.EventOperationFailed

	if data.operation.Successful() {
		actionStr = "succeeded"
		routingKey = domain.EventOperationSucceeded
	}

	actor := &hmail.Actor{
		DisplayName: "An operation",
	}

	action := &hmail.Action{
		DisplayName: actionStr,
	}

	object := &hmail.Object{
		DisplayName: fmt.Sprintf("%s/%s", data.project.Name, data.job.Name),
		Uri:         fmt.Sprintf("#/a/operations/%s", data.operation.Uuid),
	}

	finishedAt := time.Now()
	if then := data.operation.FinishedAt; then != nil {
		finishedAt = *then
	}

	subject := fmt.Sprintf("%s/%s %s on %s",
		data.project.Name, data.job.Name,
		action.DisplayName,
		finishedAt.Format(time.Stamp),
	)

	for _, user := range data.recipients {
		recipient := &hmail.Recipient{
			DisplayName: user.Name,
			Subject:     subject,
			UrlHost:     user.UrlHost,
		}

		mails = append(mails, &hmail.Mail{
			RoutingKey: routingKey,
			To:         []string{user.Email},
			Headers: textproto.MIMEHeader{
				"Subject": []string{subject},
			},
			Data: &hmail.MailContext{
				Actor:     actor,
				Action:    action,
				Object:    object,
				Recipient: recipient,
			},
		})
	}

	return mails, nil
}

func setUpOutboundChannel(log logger.Logger, conn *amqp.Connection) *amqp.Channel {
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal().Msgf("outboundchannel.open: ", err)
	}

	if err := channel.ExchangeDeclare(
		outputQueueName, // name
		"topic",         // kind
		true,            // durable
		false,           // autoDelete
		false,           // internal
		false,           // noWait
		nil,             // args
	); err != nil {
		log.Fatal().Msgf("outboundchannel.exchangedeclare: ", err)
	}

	if _, err := channel.QueueDeclare(
		outputQueueName, // queueName
		true,            // durable
		false,           // autoDelete
		false,           // exclusive
		false,           // noWait
		nil,             // args
	); err != nil {
		log.Fatal().Msgf("outboundchannel.queuedeclare: ", err)
	}

	return channel
}

//
// Dial AMQP
//
func dial(c *config.Config) *amqp.Connection {
	for {
		connection, err := amqp.Dial(c.AmqpConnectionString())
		if err != nil {
			log.Error().Msgf("error dialing amqp: %s\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Info().Msg("connected to amqp")
		return connection
	}
}

const ProgramName = "mail-dispatcher"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	c := config.GetConfig()
	fromAddress := c.MailConfig().FromAddress

	var err error

	db, err := c.DB()
	if err != nil {
		log.Fatal().Msgf("sqlx.connect:", err)
	}

	//
	// Dial AMQP
	//
	amqpConn := dial(c)
	defer amqpConn.Close()

	outboundChannel := setUpOutboundChannel(log, amqpConn)

	bus := broadcast.NewAMQPTransport(c.AmqpConnectionString(), "mail-dispatcher")
	defer bus.Close()
	work, err := bus.Consume(broadcast.Create)
	if err != nil {
		log.Fatal().Msgf("bus.consume(broadcast.create): %s", err)
	}

	for created := range work {
	retry:
		for {
			err := handleWork(log, fromAddress, c, db, outboundChannel, created)
			switch e := err.(type) {
			case *net.OpError, *amqp.Error:
				log.Error().Msgf("handleWork(): *net.OpError encountered, redialing: %s", e)
				dial(c)
				defer amqpConn.Close()
				outboundChannel = setUpOutboundChannel(log, amqpConn)
				log.Error().Msgf("Redialing successful")
				continue retry
			case nil:
				break retry
			default:
				log.Error().Msgf("handleWork: %s", e)
				break retry
			}
		}
	}
}
