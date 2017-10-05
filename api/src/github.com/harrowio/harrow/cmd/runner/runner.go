package runner

import (
	"net/url"
	"strings"
	"time"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type state string

type Runnable interface {
	Next() (*domain.Operation, error)
}

type Runner struct {
	// The runnable of operations to run
	source Runnable

	// Storage
	activitySink activity.Sink
	db           *sqlx.DB

	// externally set fields
	config   *config.Config
	interval int
	errs     chan error
	log      logger.Logger

	// internally set
	connURL *url.URL
	lxd     *LXD

	// internal management channels/etc
	stopper      chan chan bool
	healthTicker *time.Ticker

	// state (used for metrics and expvars)
	stateChange chan state
	state       state
}

func (r *Runner) Start() {

	var (
		healthChecks <-chan time.Time = make(chan time.Time)
		err          error
	)

	r.lxd = &LXD{
		config:  r.config,
		connURL: r.connURL,
		log:     r.log,
	}

	r.db, err = r.config.DB()
	if err != nil {
		r.errs <- errors.Wrap(err, "can't dial db in runner")
	}

	r.stopper = make(chan chan bool)
	r.stateChange = make(chan state)

	connectionLost := make(chan error)
	containerNetworkUp := make(chan error)
	operationPending := make(chan *domain.Operation)
	go r.MakeContainer(uuidhelper.MustNewV4(), containerNetworkUp)

	for {
		select {

		// we have a pending job, we should only recieve on this channel after we
		// successfully start a container as the Runnable fetcher only runs once we
		// get a message on the containerNetworkUp channel. The runOperation
		// goroutine will send something (maybe nil) on the err channel upon
		// completion. This gives whoever started us chance to maybe start again
		// incase we "err" out successfully
		case op := <-operationPending:
			operationPending = nil
			go r.runOperation(op)

		// were we able to start a container?
		case res := <-containerNetworkUp:

			// if starting a container failed proporgate that
			if res != nil {
				r.errs <- res
			}

			// let the log know our container is up
			r.log.Info().Msg("initial container health check passed, entering wait mode")

			// setup a ticker for the health checks
			interval := time.Duration(r.interval) * time.Second
			r.healthTicker = time.NewTicker(interval)
			healthChecks = r.healthTicker.C
			r.log.Info().Msgf("health check interval is %s", interval)

			pgDSN, err := r.config.PgDataSourceName()
			if err != nil {
				r.errs <- errors.Wrap(err, "couldn't get postgresql dsn")
			}

			go r.lxd.MaintainConnection(connectionLost)
			go func(sendOn chan<- *domain.Operation, db *sqlx.DB) {
				opdob := OperationFromDbOrBus{
					db:        db,
					log:       r.log,
					dbConnStr: pgDSN,
				}
				if err := opdob.NextOn(sendOn); err != nil {
					r.errs <- err
				}
			}(operationPending, r.db)

		// Incase we got a stop signal break this loop
		case stopped := <-r.stopper:
			r.db.Close()
			if r.healthTicker != nil {
				r.healthTicker.Stop()
			}
			if err := r.lxd.DestroyContainer(); err != nil {
				r.log.Error().Msgf("error destroying container: %s", err)
			}
			stopped <- true
			return

		// healchChecks might be a channel that never yields (if we never go into
		// <-containerNetworkUp), but then we should exit with an error immediately
		case err := <-connectionLost:
			r.errs <- errors.Wrap(err, "long running container connection lost")

		// healchChecks might be a channel that never yields (if we never go into
		// <-containerNetworkUp), but then we should exit with an error immediately
		case <-healthChecks:
			if err := r.lxd.CheckContainerNetworking(); err != nil {
				r.errs <- errors.Wrap(err, "container failed periodic networking health check")
			}
		}
	}

}

func (r *Runner) Stop() {
	r.log.Info().Msg("sending stop signal on stopper channel")
	stopped := make(chan bool)
	r.stopper <- stopped
	r.log.Info().Msg("waiting for clean shutdown...")
	<-stopped
	r.log.Info().Msg("got clean shutdown notice")
}

func (r *Runner) SetLXDConnStr(connStr string) (err error) {
	r.connURL, err = url.Parse(connStr)
	if err != nil {
		return errors.Wrap(err, "error parsing connection string as URL")
	}
	return nil
}

func (r *Runner) MakeContainer(uuid string, res chan error) {

	r.lxd.containerUuid = uuid

	r.log.Debug().Msg("checking container health")
	exists, err := r.lxd.ContainerExists()
	if !exists {
		r.log.Debug().Msg("container does not exist")
	} else {
		r.log.Debug().Msg("container exists")
	}
	if err != nil {
		res <- err
	}
	if !exists {
		if err := r.lxd.MakeContainer(); err != nil {
			res <- err
		}
		res <- r.lxd.WaitForContainerNetworking(5 * time.Minute)
	}

}

func (r *Runner) runOperation(op *domain.Operation) {
	r.log.Info().Msgf("received operation to run: %s", op.Uuid)
	o := Operation{
		op:     op,
		db:     r.db,
		config: r.config,
		lxd:    r.lxd,
		log:    r.log,
	}
	switch op.IsUserJob() {
	case true:
		r.log.Info().Msg("operation is a user job (will run in lxd)")
		o.RunOnLXDHost()
	case false:
		notifierType := strings.Split(*op.NotifierType, "_")[0]
		r.log.Info().Msg("operation is a notifier job (will run in the local shell)")
		o.RunLocally(notifierType)
	}
	r.errs <- nil
}
