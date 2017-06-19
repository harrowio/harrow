package ws

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/bus/logevent"
	"github.com/harrowio/harrow/config"
)

const (
	StatusOK                  uint32 = 4200
	StatusInternalServerError uint32 = 4500
	ProgramName                      = "ws"
)

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {

	c := config.GetConfig()

	db, err := c.DB()
	if err != nil {
		log.Fatal().Msgf("error opening database handle:", err)
	}
	defer db.Close()

	client := redis.NewTCPClient(c.RedisConnOpts(1))
	defer client.Close()
	ss := stores.NewRedisSecretKeyValueStore(client)

	w := newWs(c, db, ss)
	w.SetLogger(log)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	httpErrors := make(chan error)
	go func() {
		handler := w.newHandler()
		httpErrors <- http.ListenAndServe(c.HttpConfig().WebSocket(), handler)
	}()

	select {
	case sig := <-signals:
		log.Info().Msgf("Received signal '%s', shutting down\n", sig)
		w.shutdown(0)
	case err := <-httpErrors:
		log.Error().Msgf("HTTP error: '%s'\n", err)
		w.shutdown(1)
	}
}

type ws struct {
	config *config.Config
	db     *sqlx.DB
	ss     stores.SecretKeyValueStore
	log    logger.Logger

	additions  chan *socket
	deletions  chan *socket
	terminator chan chan error
}

func newWs(config *config.Config, db *sqlx.DB, ss stores.SecretKeyValueStore) *ws {
	ws := &ws{
		config:     config,
		db:         db,
		ss:         ss,
		additions:  make(chan *socket),
		deletions:  make(chan *socket),
		terminator: make(chan chan error),
	}
	go ws.run()
	return ws
}

func (self *ws) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *ws) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *ws) run() {
	sockets := make(map[string]*socket)
	for {
		select {
		case socket := <-self.additions:
			sockets[socket.session.ID()] = socket
		case socket := <-self.deletions:
			delete(sockets, socket.session.ID())
		case errors := <-self.terminator:
			nErrors := 0
			for _, socket := range sockets {
				socketErrors := make(chan error)
				socket.terminator <- socketErrors
				err := <-socketErrors
				if err != nil {
					nErrors++
					log.Warn().Msgf("Error shutting down Socket(%s): %s\n", socket.session.ID(), err)
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
		}
	}
}

func (self *ws) newHandler() http.Handler {
	return sockjs.NewHandler("/ws", sockjs.DefaultOptions, self.sessionHandler)
}

func (self *ws) newLogeventSource() logevent.Source {
	redisClient := redis.NewTCPClient(self.config.RedisConnOpts(0))
	ds := NewDualSource(redisClient, self.config)
	ds.SetLogger(log)
	return ds
}

func (self *ws) newBroadcastSource(name string) broadcast.Source {
	return broadcast.NewAutoDeletingAMQPTransport(self.config.AmqpConnectionString(), fmt.Sprintf("ws-%s", name))
}

func (self *ws) sessionHandler(session sockjs.Session) {
	socket := newSocket(self, session)
	socket.SetLogger(self.log)
	self.additions <- socket
	socket.run()
}

func (self *ws) remove(socket *socket) {
	self.deletions <- socket
}

func (self *ws) shutdown(status int) {
	errors := make(chan error)
	self.terminator <- errors
	<-errors
	log.Info().Msgf("Shutdown complete, exit %d", status)
	os.Exit(status)
}
