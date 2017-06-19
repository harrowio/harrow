package notifier

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

type DbNotificationRules struct {
	db *sqlx.DB
}

func NewDbNotificationRules(db *sqlx.DB) *DbNotificationRules {
	return &DbNotificationRules{
		db: db,
	}
}

func (self *DbNotificationRules) FindByProjectUuid(projectUuid string) ([]*domain.NotificationRule, error) {
	tx, err := self.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	store := stores.NewDbNotificationRuleStore(tx)
	return store.FindByProjectUuid(projectUuid)
}
