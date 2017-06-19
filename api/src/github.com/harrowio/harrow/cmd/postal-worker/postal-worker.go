//go:generate go-bindata -pkg postalWorker -o compiled-templates.go  -ignore (^\..*$)|(^.*/\..*$) -prefix ../../mail ../../mail/...
package postalWorker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/hmail"
	"github.com/rs/zerolog"

	"github.com/mohamedattahri/mail"
	"github.com/streadway/amqp"
)

func emailNameFromRoutingKey(routingKey string) string {
	return strings.Replace(routingKey, ".", "/", -1)
}

func newMail(emailName string, hMail *hmail.Mail) (*mail.Message, error) {
	files, err := AssetDir(emailName)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("newMail: no files in %q", emailName)
	}

	msg := mail.NewMessage()

	// Set From: header
	from, err := mail.ParseAddress(hMail.From)
	if err != nil {
		log.Error().Msgf("error parsing from: header", hMail.From)
		return nil, err
	}
	msg.SetFrom(from)

	// Set To: headers
	for _, to := range hMail.To {
		to, err := mail.ParseAddress(to)
		if err != nil {
			log.Error().Msgf("error parsing to: headers")
			return nil, err
		}
		msg.To().Add(to)
	}

	// Copy the headers from the harrow.Mail
	for header, values := range hMail.Headers {
		for _, value := range values {
			textproto.MIMEHeader(msg.Header).Add(header, value)
		}
	}

	multipart := mail.NewMultipart("multipart/alternative", msg)

	// Add parts from templates (html, text)
	for _, filename := range files {
		if strings.Contains(filename, ".tmpl") {
			tplContent, err := Asset(filepath.Join(emailName, filename))
			if err != nil {
				return nil, err
			}
			base := filepath.Base(filename)
			// remove extension
			base = base[0 : len(base)-len(filepath.Ext(base))]
			mimetype := mime.TypeByExtension("." + base)
			if mimetype == "" {
				return nil, fmt.Errorf("Could not determine mimetype for %q", base)
			}
			tpl, err := template.New(filename).Parse(string(tplContent))
			if err != nil {
				return nil, err
			}
			buf := new(bytes.Buffer)
			tpl.Execute(buf, hMail.Data)
			multipart.AddText(mimetype, buf)
		}
	}

	// Add global attachments
	err = addAttachments(multipart, filepath.Join("global", "attachments"))
	if err != nil {
		return nil, err
	}
	// Add mail-specific attachments
	err = addAttachments(multipart, filepath.Join(emailName, "attachments"))
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Add attachments from the given folder
// These can be referred to with the url cid:<%-encoded bassename of the attachment>
// for example: <img src="cid:logo.png">.
// Attachments are always refered to by basename in the mail.
func addAttachments(mp *mail.Multipart, dirname string) error {
	attachmentFiles, err := AssetDir(dirname)
	if err == nil && len(attachmentFiles) > 0 {
		for _, filename := range attachmentFiles {
			attachment, err := Asset(filepath.Join(dirname, filename))
			if err != nil {
				return err
			}
			err = mp.AddAttachment("inline", filepath.Base(filename), "" /* auto-mimetype */, bytes.NewReader(attachment))
			if err != nil {
				return err
			}
		}
	} // Ignore attachments dir absent or empy cases
	return nil
}

func handleWork(d amqp.Delivery) {
	name := emailNameFromRoutingKey(d.RoutingKey)

	data := hmail.Mail{}
	if err := json.Unmarshal(d.Body, &data); err != nil {
		log.Warn().Msgf("error unmarshaling body: %s\n", err)
		d.Nack(false, false)
		return
	}

	log.Debug().Msgf("mail data is: %#v", data)

	msg, err := newMail(name, &data)
	if err != nil {
		log.Warn().Msgf("failed to send email: %s\n", err)
		d.Nack(false, false)
		return
	}

	if err := sendMailWithSendmail(msg); err != nil {
		log.Info().Msgf("error sending with sendmail %s", err)
		d.Nack(false, false)
		return
	}

	d.Ack(false)
}

func sendMailWithSendmail(msg *mail.Message) error {
	al := msg.To()
	addresses, err := al.Addresses()
	if err != nil {
		return err
	}
	for _, to := range addresses {
		cmd := &exec.Cmd{
			Stdin: bytes.NewReader(msg.Bytes()),
			Args:  []string{"/usr/sbin/sendmail", "-f", msg.From().Address, to.Address},
			Path:  "/usr/sbin/sendmail",
		}

		log.Debug().Msgf("sent mail: from: %q to: %q subject: %q", msg.From(), to, msg.Subject())

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

const ProgramName = "postal-worker"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	c := config.GetConfig()

	work, connection, err := dial(c)
	defer connection.Close()
	if err != nil {
		log.Fatal().Msgf("dial: %s", err)
	}
	for {
		d, ok := <-work
		if !ok {
			log.Warn().Msg("lost amqp connection, re-dialing")
			work, connection, err = dial(c)
			defer connection.Close()
			if err != nil {
				panic(fmt.Sprintf("Unable to open work channel: %s", err))
			}
			log.Warn().Msg("successfully reconnected")
			continue

		}
		handleWork(d)
	}
}

func dial(conf *config.Config) (<-chan amqp.Delivery, *amqp.Connection, error) {
	//
	// Dial AMQP
	//
	connection := (*amqp.Connection)(nil)
	err := (error)(nil)
	for {
		connection, err = amqp.Dial(conf.AmqpConnectionString())
		if err != nil {
			log.Error().Msgf("error dialing amqp: %s\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Info().Msg("connected to amqp")
		break
	}

	//
	// AMQP Channel
	//
	amqpChan, err := connection.Channel()
	if err != nil {
		log.Fatal().Msgf("channel.open: %s\n", err)
	}

	//             ExchangeDeclare(name string, kind string, durable, autoDelete, internal, noWait bool, args Table) error
	err = amqpChan.ExchangeDeclare("work.mail.outbound.all", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Msgf("exchange.declare: %s\n", err)
	}

	_, err = amqpChan.QueueDeclare(
		"work.mail.outbound.all", // queueName
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		log.Fatal().Msgf("queue declare: %s\n", err)
	}

	err = amqpChan.QueueBind(
		"work.mail.outbound.all", // queueName
		"#", // routingKey
		"work.mail.outbound.all", // exchangeName
		false, // noWait
		nil,   // args
	)
	if err != nil {
		log.Fatal().Msgf("queue.bind: %s\n", err)
	}

	deliveries, err := amqpChan.Consume(
		"work.mail.outbound.all", // queueName
		"",    // consumerName
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)

	return deliveries, connection, err
}
