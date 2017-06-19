package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbSubscriptionStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbSubscriptionStore(tx *sqlx.Tx) *DbSubscriptionStore {
	return &DbSubscriptionStore{tx: tx}
}

func (store *DbSubscriptionStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbSubscriptionStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbSubscriptionStore) Create(subscription *domain.Subscription) (string, error) {

	if len(subscription.Uuid) == 0 {
		subscription.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO subscriptions
        (uuid, user_uuid, watchable_uuid, watchable_type, event_name)
        VALUES
        (:uuid, :user_uuid, :watchable_uuid, :watchable_type, :event_name)`

	_, err := store.tx.NamedExec(q, subscription)
	return subscription.Uuid, err
}

func (store *DbSubscriptionStore) Delete(subscriptionUuid string) error {

	q := `DELETE FROM subscriptions WHERE uuid = $1 AND archived_at IS NULL`
	result, err := store.tx.Exec(q, subscriptionUuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, err := result.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	} else {
		return resolveErrType(err)
	}

}

func (store *DbSubscriptionStore) Find(watchableId, event, userUuid string) (*domain.Subscription, error) {

	q := `SELECT s.* FROM subscriptions s
        WHERE watchable_uuid = $1
          AND event_name = $2
          AND user_uuid = $3
          AND archived_at IS NULL`

	result := domain.Subscription{}
	err := store.tx.Get(&result, q, watchableId, event, userUuid)
	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (store *DbSubscriptionStore) FindEventsForUser(watchableId, userUuid string) ([]string, error) {

	q := `SELECT DISTINCT event_name
        FROM subscriptions
        WHERE watchable_uuid = $1
          AND user_uuid = $2
          AND archived_at IS NULL
	`

	events := []string{}

	err := store.tx.Select(&events, q, watchableId, userUuid)
	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}
	if err != nil {
		return nil, err
	}

	return events, nil
}
