package ws

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"gopkg.in/igm/sockjs-go.v2/sockjs"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/http"
	"github.com/harrowio/harrow/logger"

	_ "net/http/pprof"
)

type socket struct {
	ws                  *ws
	session             sockjs.Session
	user                *domain.User
	terminator          chan chan error
	commandTerminations chan string
	log                 logger.Logger
}

func doNothing() {}

func newSocket(ws *ws, session sockjs.Session) *socket {
	return &socket{
		ws:                  ws,
		session:             session,
		terminator:          make(chan chan error),
		commandTerminations: make(chan string),
	}
}

func (self *socket) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *socket) SetLogger(l logger.Logger) {
	self.log = l
}

// Wrap the blocking call to Recv() in a goroutine and transform it to a
// channel api. The only condition for termination of that goroutine is
// receiving an error from Recv(), but this is an inherent problem of the
// blocking call and cannot be solved.
func (self *socket) recv() (chan string, chan error) {
	messages := make(chan string)
	errors := make(chan error)
	go func() {
		for {
			data, err := self.session.Recv()
			if err != nil {
				errors <- err
				return
			}
			messages <- data
		}
	}()
	return messages, errors
}

func (self *socket) String() string {
	return fmt.Sprintf("Socket %s for %s", self.session.ID(), self.user.Email)
}

func (self *socket) run() {
	defer func() {
		err := recover()
		if err != nil {
			self.log.Error().Msgf("Recovered from panic, terminating socket: %s", err)
			self.log.Error().Msg(string(debug.Stack()))
			self.terminate(fmt.Errorf("panic: %s", err))
		}
	}()
	messages, recvErrors := self.recv()
	running := make(map[string]Command)

	for {
		select {
		case errors := <-self.terminator:

			nErrors := 0
			for _, cmd := range running {
				err := cmd.Terminate()
				if err != nil {
					nErrors++
					self.log.Error().Msgf("Error shutting down Command(%s): %s\n", cmd.ID(), err)
				}
			}
			if errors != nil {
				if nErrors > 0 {
					errors <- fmt.Errorf("encountered %d errrors while shutting down", nErrors)
				} else {
					errors <- nil
				}
			}
			return
		case cid := <-self.commandTerminations:
			delete(running, cid)
		case err := <-recvErrors:
			self.terminate(err)
		case data := <-messages:
			peek := &baseCmd{}
			err := json.Unmarshal([]byte(data), peek)
			if err != nil {
				self.log.Error().Msgf("Unable to unmarshal Command: err='%s', json='%s'\n", err, string(data))
				continue
			}
			if peek.Stop {
				if _, ok := running[peek.CID]; ok {
					running[peek.CID].Terminate()
				}
				continue
			}
			_, exists := running[peek.CID]
			if exists {
				self.log.Error().Msgf("Command can't be run twice json='%s'\n", string(data))
				continue
			}
			cmd, err := self.newCommand(peek.Command, data)
			if err != nil {
				self.log.Error().Msgf("Unable to build Command object err='%s' json='%s'\n", err, string(data))
				continue
			}
			tx, err := self.ws.db.Beginx()
			if err != nil {
				self.log.Error().Msgf("Unable to start tx err='%s' json='%s\n'", err, string(data))
				continue
			}
			defer tx.Rollback()
			self.user, err = http.CurrentUser(self.ws.config, tx, cmd.GetSessionUuid())
			if err != nil && err != http.ErrSessionInvalidated {
				self.log.Error().Msgf("Unable to authenticate err='%s' json='%s', terminating\n", err, string(data))
				self.terminate(err)
				continue
			}
			authzService := authz.NewService(tx, self.user, self.ws.config)
			can, err := cmd.Authorize(tx, authzService)
			tx.Rollback()
			if err != nil || !can {
				self.log.Error().Msgf("Unauthorized err='%s' json='%s'\n", err, string(data))
				continue
			}
			err = cmd.Exec()
			if err != nil {
				self.log.Error().Msgf("Unable to execute Command err='%s' json='%s'\n", err, string(data))
				continue
			}
			running[peek.CID] = cmd
		}
	}
}

func (self *socket) removeCommand(cmd Command) {
	go func() {
		self.commandTerminations <- cmd.ID()
	}()
}

func (self *socket) send(data []byte) error {
	return self.session.Send(string(data))
}

func (self *socket) terminate(reason error) {
	if reason == nil {
		self.close(StatusOK, "OK")
	} else {
		self.close(StatusInternalServerError, reason.Error())
	}
	self.ws.remove(self)
	errorChan := make(chan error)
	errors := make([]string, 0, 1)
	self.terminator <- errorChan
	for err := range errorChan {
		errors = append(errors, err.Error())
	}
	self.log.Error().Msgf("error(s) terminating socket, sessionId=%s: %s", self.session.ID(), strings.Join(errors, ", "))
}

func (self *socket) close(code uint32, reason string) {
	//log.Debug().Msgf("close socket %s: %d(%s)\n", self.session.ID(), code, reason)
	// Try to close the Session.
	// N.B.: When we close because the client went away the Session is already
	// in closed state and trying to Close it again will error. As there is no
	// way to check if the Session is closed or not we try anyways and ignore
	// the error.
	_ = self.session.Close(code, reason)
}
