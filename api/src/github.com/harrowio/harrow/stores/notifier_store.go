package stores

import (
	"fmt"
	"strings"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type DbNotifierStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbNotifierStore(tx *sqlx.Tx) *DbNotifierStore {
	return &DbNotifierStore{
		tx: tx,
	}
}

func (self *DbNotifierStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbNotifierStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *DbNotifierStore) FindByUuidAndType(uuid, typename string) (interface{}, error) {
	q := fmt.Sprintf(`SELECT * FROM %q WHERE uuid = $1`, typename)
	typename = normalizeNotifierTypeName(typename)
	var dest interface{}
	switch typename {
	case "email_notifiers":
		dest = new(domain.EmailNotifier)
	case "job_notifiers":
		dest = new(domain.JobNotifier)
	case "slack_notifiers":
		dest = new(domain.SlackNotifier)
	default:
		return nil, fmt.Errorf("Unsupported notifier type: %q", typename)
	}

	if err := self.tx.Get(dest, q, uuid); err != nil {
		return nil, err
	}

	return dest, nil
}

// normalizeNotifierTypeName converts names of the form fooNotifiers
// to foo_notifiers.
func normalizeNotifierTypeName(typename string) string {
	if strings.HasSuffix(typename, "Notifiers") {
		return strings.Replace(typename, "Notifiers", "_notifiers", 1)
	} else if strings.HasSuffix(typename, "Notifier") {
		return strings.Replace(typename, "Notifier", "_notifiers", 1)
	}

	return typename
}
