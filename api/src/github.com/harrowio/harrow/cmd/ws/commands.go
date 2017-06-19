package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/bus/logevent"
	"github.com/harrowio/harrow/loxer"
	"github.com/harrowio/harrow/stores"
)

var zl zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

var BatchInterval = 500 * time.Millisecond

func (self *socket) newCommand(name, data string) (Command, error) {
	switch name {
	case "subLogevents":
		return newSubLogeventsCmd(self, data)
	case "subRow":
		return newSubRowCmd(self, data)
	default:
		return nil, fmt.Errorf("Unsupported Command: %s", name)
	}
}

type Command interface {
	ID() string
	Exec() error
	Terminate() error
	GetSessionUuid() string
	Authorize(*sqlx.Tx, authz.Service) (bool, error)
}

type baseCmd struct {
	sync.Mutex
	Command     string `json:"command"`
	CID         string `json:"cid"`
	SessionUuid string `json:"sessionUuid"`
	Stop        bool   `json:"stop"`
	socket      *socket
	terminator  chan chan error
}

func (self *baseCmd) ID() string {
	return self.CID
}

func (self *baseCmd) GetSessionUuid() string {
	return self.SessionUuid
}

func (self *baseCmd) Terminate() error {
	self.Lock()
	defer self.Unlock()
	if self.terminator == nil {
		return nil
	}
	errors := make([]string, 0, 1)
	errorChan := make(chan error)
	self.terminator <- errorChan
	for err := range errorChan {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	self.terminator = nil

	if len(errors) > 0 {
		return fmt.Errorf("error(s) while terminating, CID=%s: %s", self.CID, strings.Join(errors, ", "))
	} else {
		return nil
	}
}

type updateCmd struct {
	CID string `json:"cid"`
}

////////////////////////////////////////////////////////////////////////////////
// subLogeventsCmd
////////////////////////////////////////////////////////////////////////////////
func newBatchSender(socket *socket, cid string, interval time.Duration) (chan<- *logevent.Message, <-chan bool) {
	incoming := make(chan *logevent.Message)
	stop := make(chan bool)
	batch := make([]*logevent.Message, 0)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case i, ok := <-incoming:
				if !ok {
					ticker.Stop()
					err := writeBatch(socket, cid, batch)
					if err != nil {
						zl.Debug().Msgf("writeBatch(): %s", err)
					}
					err = writeBatch(socket, cid, []*logevent.Message{
						{E: loxer.SerializedEvent{Inner: loxer.EOF}},
					})
					if err != nil {
						zl.Debug().Msgf("writeBatch(,,loxer.EOF): %s", err)
					}
					close(stop)
					return
				}
				batch = append(batch, i)
			case <-ticker.C:
				err := writeBatch(socket, cid, batch)
				if err != nil {
					zl.Debug().Msgf("writeBatch(): %s", err)
				}
				batch = make([]*logevent.Message, 0)
			}
		}
	}()
	return incoming, stop
}

func writeBatch(socket *socket, cid string, batch []*logevent.Message) error {
	if len(batch) == 0 {
		return nil
	}

	update := &logeventUpdate{
		updateCmd: updateCmd{
			CID: cid,
		},
		Logevents: batch,
	}

	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("unable to marshal updateCmd ws=%s: %s", socket.session.ID(), err)
	}
	err = socket.send(data)
	if err != nil {
		return fmt.Errorf("unable to send data ws=%s: %s", socket.session.ID(), err)
	}
	return nil
}

type subLogeventsCmd struct {
	baseCmd
	OperationUuid string `json:"operationUuid"`
}

type logeventUpdate struct {
	updateCmd
	Logevents []*logevent.Message `json:"logevents"`
}

func newSubLogeventsCmd(socket *socket, data string) (*subLogeventsCmd, error) {
	c := &subLogeventsCmd{}
	c.terminator = make(chan chan error)
	c.socket = socket
	err := json.Unmarshal([]byte(data), c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (self *subLogeventsCmd) String() string {
	return fmt.Sprintf("subLogeventsCmd(%s)", self.OperationUuid)
}

func (self *subLogeventsCmd) Exec() error {
	logeventSource := self.socket.ws.newLogeventSource()
	messages, err := logeventSource.Consume(self.OperationUuid)
	if err != nil {
		logeventSource.Close()
		return err
	}
	batch, batchStop := newBatchSender(self.socket, self.CID, BatchInterval)
	go func() {
		for {
			select {
			case errors := <-self.terminator:
				close(batch)
				// wait until batchSender is closed
				<-batchStop
				closeRes := logeventSource.Close()
				if errors != nil {
					errors <- closeRes
					close(errors)
				}
				self.socket.removeCommand(self)
				return
			case msg, ok := <-messages:
				if !ok {
					messages = nil
					go self.Terminate()
					continue
				}
				batch <- msg
				//         &logevent.Message{
				// 	FD: msg.FD,
				// 	T:  msg.T,
				// 	E:  loxer.SerializedEvent{Inner: msg.Event()},
				// }
			}
		}
	}()
	return nil
}

func (self *subLogeventsCmd) Authorize(tx *sqlx.Tx, service authz.Service) (bool, error) {
	store := stores.NewDbOperationStore(tx)
	operation, err := store.FindByUuid(self.OperationUuid)
	if err != nil {
		return false, fmt.Errorf("Unable to load operation %s: %s", self.OperationUuid, err)
	}
	return service.CanRead(operation)
}

////////////////////////////////////////////////////////////////////////////////
// subRowCmd
////////////////////////////////////////////////////////////////////////////////
type subRowCmd struct {
	baseCmd
	Table string `json:"table"`
	Uuid  string `json:"uuid"`
}

type rowUpdate updateCmd

func newSubRowCmd(socket *socket, data string) (*subRowCmd, error) {
	c := &subRowCmd{}
	c.terminator = make(chan chan error)
	c.socket = socket
	err := json.Unmarshal([]byte(data), c)
	zl.Debug().Msgf("subcribed to %s %s", c.Table, c.Uuid)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (self *subRowCmd) String() string {
	return fmt.Sprintf("subRowCmd(%s,%s)", self.Table, self.Uuid)
}

func (self *subRowCmd) Exec() error {

	source := self.socket.ws.newBroadcastSource(self.CID)
	changes, err := source.Consume(broadcast.Change)
	zl.Debug().Msgf("cid=%s table=%s uuid=%s subscribed to AMQP: err=%s", self.CID, self.Table, self.Uuid, err)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case errors := <-self.terminator:
				closeRes := source.Close()
				if errors != nil {
					errors <- closeRes
					close(errors)
				}
				self.socket.removeCommand(self)
				return
			case change, ok := <-changes:
				if !ok {
					go self.Terminate()
					continue
				}
				zl.Debug().Msgf("%s change want=%s got=%s %s", self.Uuid, self.Table, change.Table(), change.UUID())
				// discard updates we are not interested in
				// TODO: remove this check when broadcast.Source gains the ability to
				// subscribe to Table/UUID pairs directly.
				if change.Table() != self.Table || change.UUID() != self.Uuid {
					change.RejectForever()
					continue
				}

				err := self.sendUpdate()

				if err != nil {
					zl.Info().Msgf("Unable to send update, terminating: %s\n", err)
					go self.Terminate()
					change.RejectForever()
					continue
				}

				change.Acknowledge()
			}
		}
	}()
	return nil
}

func (self *subRowCmd) sendUpdate() error {
	update := &rowUpdate{
		CID: self.CID,
	}
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}
	return self.socket.send(data)
	if err != nil {
		return err
	}
	return nil
}

func (self *subRowCmd) Authorize(tx *sqlx.Tx, service authz.Service) (bool, error) {
	s, err := self.getSubject(tx)
	if err != nil {
		return false, fmt.Errorf("unable to load subject: %s", err)
	}
	return service.CanRead(s)
}

func (self *subRowCmd) getSubject(tx *sqlx.Tx) (domain.Subject, error) {
	switch self.Table {
	case "deliveries":
		store := stores.NewDbDeliveryStore(tx)
		return store.FindByUuid(self.Uuid)
	case "environments":
		store := stores.NewDbEnvironmentStore(tx)
		return store.FindByUuid(self.Uuid)
	case "invitations":
		store := stores.NewDbInvitationStore(tx)
		return store.FindByUuid(self.Uuid)
	case "jobs":
		store := stores.NewDbJobStore(tx)
		return store.FindByUuid(self.Uuid)
	case "oauth_tokens":
		store := stores.NewDbOAuthTokenStore(tx)
		return store.FindByUuid(self.Uuid)
	case "operations":
		store := stores.NewDbOperationStore(tx)
		return store.FindByUuid(self.Uuid)
	case "organization_memberships":
		return nil, errors.New("DbOrganizationMembershipStore.FindByUuid not defined")
	case "organizations":
		store := stores.NewDbOrganizationStore(tx)
		return store.FindByUuid(self.Uuid)
	case "project_memberships":
		return nil, errors.New("DbProjectMembershipStore.FindByUuid not defined")
	case "projects":
		store := stores.NewDbProjectStore(tx)
		return store.FindByUuid(self.Uuid)
	case "repositories":
		store := stores.NewDbRepositoryStore(tx)
		return store.FindByUuid(self.Uuid)
	case "repository_credentials":
		store := stores.NewRepositoryCredentialStore(self.socket.ws.ss, tx)
		return store.FindByUuid(self.Uuid)
	case "schedules":
		store := stores.NewDbScheduleStore(tx)
		return store.FindByUuid(self.Uuid)
	case "secrets":
		store := stores.NewSecretStore(self.socket.ws.ss, tx)
		return store.FindByUuid(self.Uuid)
	case "sessions":
		store := stores.NewDbSessionStore(tx)
		return store.FindByUuid(self.Uuid)
	case "subscriptions":
		return nil, errors.New("DbSubscriptionStore.FindByUuid not defined")
	case "targets":
		store := stores.NewDbTargetStore(tx)
		return store.FindByUuid(self.Uuid)
	case "tasks":
		store := stores.NewDbTaskStore(tx)
		return store.FindByUuid(self.Uuid)
	case "user_blocks":
		return nil, errors.New("DbUserBlockStore.FindByUuid not defined")
	case "users":
		store := stores.NewDbUserStore(tx, self.socket.ws.config)
		return store.FindByUuid(self.Uuid)
	case "webhooks":
		store := stores.NewDbWebhookStore(tx)
		return store.FindByUuid(self.Uuid)
	case "workspace_base_images":
		return nil, errors.New("WorkspaceBaseImage is no domain.Subjetc")
	}
	return nil, fmt.Errorf("unknown table %s", self.Table)
}
