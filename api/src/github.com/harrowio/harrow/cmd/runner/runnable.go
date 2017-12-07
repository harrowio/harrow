package runner

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/harrowio/harrow/cast"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type OperationFromDbOrBus struct {
	dbConnStr string
	db        *sqlx.DB
	log       logger.Logger
}

func (ofdob *OperationFromDbOrBus) WaitForNew(quit chan bool) bool { // look again
	ofdob.log.Info().Msg("waiting for sth to happen on db new-operation channel")
	l := pq.NewListener(ofdob.dbConnStr, 10*time.Second, time.Minute, nil)
	if err := l.Listen("new-operation"); err != nil {
		ofdob.log.Fatal().Msgf("error listening on pg channel %q: %s", "new-operation", err)
	}
	defer l.Close()
	select {
	case <-l.Notify:
		ofdob.log.Info().Msg("something happened in db, returning")
		return true // look again
	case <-quit:
		ofdob.log.Info().Msg("searcher goroutine got kill sig, closing listener and returning")
		l.Close()
		return false // don't look again
	}
}

// appendStatusLog appends a record to the stauts log for the event given
// a transaction to use. The transaction must be given as we have two
// different transactions in play here. If nil is given then a new one will
// be started on the OperationFromDbOrBus.db
func appendStatusLog(log logger.Logger, tx *sqlx.Tx, uuid, entryType, subject string) error {

	operationStore := stores.NewDbOperationStore(tx)
	operation, err := operationStore.FindByUuid(uuid)
	if err != nil {
		return errors.Wrap(err, "error looking up operation")
	}

	entry := cast.NewStatusLogEntry(entryType, subject)
	operation.HandleEvent(entry.Payload)

	if err := operationStore.MarkStatusLogs(uuid, operation.StatusLogs); err != nil {
		return errors.Wrap(err, "error appending status log")
	}

	return nil
}

// Next on can immediately return an error, else it will eventually send
// an operation when one becomes available on the channel given
func (ofdob *OperationFromDbOrBus) NextOn(quit chan bool, ch chan<- *domain.Operation) error {
	op, err := ofdob.Next()
	if err != nil {
		return err
	}
	if op != nil {
		ch <- op
		return nil
	}
	if lookAgain := ofdob.WaitForNew(quit); lookAgain {
		return ofdob.NextOn(quit, ch)
	} else {
		return nil
	}
}

// Next uses it's own transaction to atomically select the next unstarted operation
// from the database. It uses a single transaction to get the lock, as if this transaction
// would be shared for the status message updates ("waiting for vm...", etc) then the
// status messages would be delayed until the end of the operation.
//
// The method will call itself if it retrieves an outdated operation, after marking
// the initially selected operation as timed out. (Blue in the UI, probably)
//
// The locks used here are advisory, database sessions without a transaction may still
// be able to get a handle on this row and update other fields.
func (ofdob *OperationFromDbOrBus) Next() (*domain.Operation, error) {

	fmt.Println("entering OperationFromDbOrBus.Next()")
	tx, err := ofdob.db.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "could not start database transaction")
	}
	defer tx.Commit()

	ofdob.log.Info().Msg("getting next unstarted operation from database")

	var ops []domain.Operation = []domain.Operation{}
	var opStore *stores.DbOperationStore = stores.NewDbOperationStore(tx)

	// started_at is our only "start" field
	// and the other five are "stop" fields
	// we're looking for anything unstarted that hasn't been stopped
	// for any reason.
	query := ` SELECT * FROM operations WHERE (started_at IS NULL) AND (canceled_at IS NULL AND timed_out_at IS NULL AND failed_at IS NULL AND finished_at IS NULL AND archived_at IS NULL) ORDER BY created_at ASC;`
	err = tx.Select(&ops, query)
	if err == sql.ErrNoRows {
		ofdob.log.Debug().Msg("no rows found, but no errors")
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not select next unstarted operation from database")
	}

	ofdob.log.Info().Int("unstarted_operations", len(ops)).Msg("unstarted operations")

	// The operations are sorted by the time at which they were created
	// which means we should start with the oldest and try and lock it
	var op *domain.Operation
	for _, o := range ops {
		var gotLock bool
		var query string = fmt.Sprintf(`SELECT pg_try_advisory_lock(('x' || lpad('%s', 8, '0'))::bit(32)::int);`, o.Uuid)
		err := tx.Get(&gotLock, query)
		if err != nil {
			return nil, errors.Wrap(err, "error attempting lock")
		}
		if gotLock == true {
			ofdob.log.Debug().Msgf("got lock, %#v", o)
			op = &o
			break
		} else {
			ofdob.log.Debug().Msg("failed to get lock, trying again")
		}
	}

	// Did we exit the loop because of a successful lock or
	if op == nil {
		ofdob.log.Info().Msg("failed to lock a record, returning for another shot on the next tick")
		return nil, nil
	} else {
		ofdob.log.Info().Msgf("locked operation uuid %s, proceeding", op.Uuid)
	}

	// TODO: risky, pointer dereference for a possibly nil field? (ttl calc and age to domain.Operation)
	ofdob.log.Info().Msgf("found operation %s (age: %s), checking ttl", op.Uuid, time.Now().UTC().Sub(*op.CreatedAt))

	if op.CreatedAt.Add(time.Duration(op.TimeLimit) * time.Second).Before(time.Now().UTC()) {
		ofdob.log.Info().Msg("operation has exceeded ttl, status will be updated and marked as timed out")
		if err := appendStatusLog(ofdob.log, tx, op.Uuid, "ttl.expired", fmt.Sprintf("failed to start before the %s time limit expired", time.Duration(op.TimeLimit)*time.Second)); err != nil {
			return nil, errors.Wrap(err, "could not append ttl.expired message to operation status logs")
		}
		if err := opStore.MarkAsTimedOut(op.Uuid); err != nil {
			return nil, errors.Wrap(err, "could not mark expired operation as timed out")
		}
		tx.Commit()
		ofdob.log.Info().Str("runnable", "Next()").Msg("recursing")
		return ofdob.Next()
	}

	if err := appendStatusLog(ofdob.log, tx, op.Uuid, "vm.reserved", fmt.Sprintf("Reserved, will be	started (wait time %s)", time.Now().UTC().Sub(*op.CreatedAt))); err != nil {
		return nil, errors.Wrap(err, "could not append vm.reserved message to operation status logs")
	}

	ofdob.log.Info().Str("runnable", "Next()").Msg("returning")
	return op, nil
}
