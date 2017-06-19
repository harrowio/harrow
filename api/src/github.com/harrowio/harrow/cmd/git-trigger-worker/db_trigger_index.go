package gitTriggerWorker

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

type DbTriggerIndex struct {
	db     *sqlx.DB
	initTx func(*sqlx.Tx) error
}

func NewDbTriggerIndex(db *sqlx.DB) *DbTriggerIndex {
	return &DbTriggerIndex{
		db:     db,
		initTx: func(*sqlx.Tx) error { return nil },
	}
}

// FindTriggersForActivity returns all Git triggers found in the
// database for the project that is associated with activity.
func (self *DbTriggerIndex) FindTriggersForActivity(activity *domain.Activity) ([]*domain.GitTrigger, error) {
	tx, err := self.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := self.initTx(tx); err != nil {
		return nil, err
	}

	projectUuid := activity.ProjectUuid()
	if projectUuid == "" {
		return nil, nil
	}

	triggerStore := stores.NewDbGitTriggerStore(tx)
	return triggerStore.FindByProjectUuid(projectUuid)
}

// InitTxWith registers fn to run with the current transaction after a
// new transaction has been acquired.
//
// This method is introduced to inject test data into the transaction
// in the unit tests.
func (self *DbTriggerIndex) InitTxWith(fn func(tx *sqlx.Tx) error) *DbTriggerIndex {
	self.initTx = fn
	return self
}
