package notifier

import (
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbScheduler struct {
	db *sqlx.DB
}

func NewDbScheduler(db *sqlx.DB) *DbScheduler {
	return &DbScheduler{
		db: db,
	}
}

func (self *DbScheduler) ScheduleNotification(rule *domain.NotificationRule, activity *domain.Activity) error {
	tx, err := self.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	params := domain.NewOperationParameters()
	params.Reason = domain.OperationTriggeredByNotificationRule
	params.TriggeredByNotificationRule = rule.Uuid
	params.TriggeredByActivityId = activity.Id

	now := time.Now()
	operation := &domain.Operation{
		Uuid:                   uuidhelper.MustNewV4(),
		TimeLimit:              300,
		Type:                   domain.OperationTypeNotifierInvoke,
		NotifierUuid:           &rule.NotifierUuid,
		NotifierType:           &rule.NotifierType,
		WorkspaceBaseImageUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007",
		CreatedAt:              &now,
		Parameters:             params,
	}

	if _, err := stores.NewDbOperationStore(tx).Create(operation); err != nil {
		return err
	}

	return tx.Commit()
}
