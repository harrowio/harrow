package stores

import (
	"database/sql"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbDeliveryStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbDeliveryStore(tx *sqlx.Tx) *DbDeliveryStore {
	return &DbDeliveryStore{tx: tx}
}

func (store *DbDeliveryStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbDeliveryStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbDeliveryStore) Create(delivery *domain.Delivery) (string, error) {

	return store.CreateAt(delivery, time.Now())
}

func (store *DbDeliveryStore) CreateAt(delivery *domain.Delivery, at time.Time) (string, error) {

	if delivery.Uuid == "" {
		delivery.Uuid = uuidhelper.MustNewV4()
	}
	delivery.DeliveredAt = at

	q := `INSERT INTO deliveries (
		uuid,
		webhook_uuid,
		request,
		delivered_at,
		schedule_uuid
	) VALUES (
		:uuid,
		:webhook_uuid,
		:request,
		:delivered_at,
		:schedule_uuid
	);`

	_, err := store.tx.NamedExec(q, delivery)
	if err != nil {
		return "", resolveErrType(err)
	}

	return delivery.Uuid, nil
}
func (store *DbDeliveryStore) FindByUuid(uuid string) (*domain.Delivery, error) {

	q := `SELECT * FROM deliveries WHERE uuid = $1 AND archived_at IS NULL`
	result := &domain.Delivery{}

	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}
	if err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store *DbDeliveryStore) FindByWebhookUuid(uuid string) ([]*domain.Delivery, error) {

	q := `SELECT * FROM deliveries WHERE webhook_uuid = $1 AND archived_at IS NULL ORDER BY delivered_at DESC LIMIT 20`
	result := []*domain.Delivery{}
	err := store.tx.Select(&result, q, uuid)

	return result, err
}
