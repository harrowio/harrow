package gitTriggerWorker

import (
	"fmt"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

type DbScheduler struct {
	db       *sqlx.DB
	finishTx func(*sqlx.Tx) error
	initTx   func(*sqlx.Tx) error
}

func NewDbScheduler(db *sqlx.DB) *DbScheduler {
	return &DbScheduler{
		db: db,
		finishTx: func(*sqlx.Tx) error {
			return nil
		},
		initTx: func(*sqlx.Tx) error {
			return nil
		},
	}
}

// FinishTxWith registers fn to run with the current transaction after a
// new transaction has been acquired.
//
// This method is introduced to inject test data into the transaction
// in the unit tests.
func (self *DbScheduler) FinishTxWith(fn func(*sqlx.Tx) error) *DbScheduler {
	self.finishTx = fn
	return self
}

// InitTxWith registers fn to run with the current transaction after a
// new transaction has been acquired.
//
// This method is introduced to inject test data into the transaction
// in the unit tests.
func (self *DbScheduler) InitTxWith(fn func(*sqlx.Tx) error) *DbScheduler {
	self.initTx = fn
	return self
}

func (self *DbScheduler) ScheduleJob(forTrigger *domain.GitTrigger, params *domain.OperationParameters) error {
	tx, err := self.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := self.initTx(tx); err != nil {
		return err
	}

	now := "now"
	params.Reason = domain.OperationTriggeredByGitTrigger
	params.TriggeredByGitTrigger = forTrigger.Uuid
	params.GitTriggerName = forTrigger.Name
	schedule := &domain.Schedule{
		UserUuid:    forTrigger.CreatorUuid,
		JobUuid:     forTrigger.JobUuid,
		Description: fmt.Sprintf("Triggered by git-trigger"),
		CreatedAt:   time.Now(),
		Timespec:    &now,
		Parameters:  params,
	}

	if _, err := stores.NewDbScheduleStore(tx).Create(schedule); err != nil {
		return err
	}

	if err := self.finishTx(tx); err != nil {
		return err
	}

	return tx.Commit()
}
