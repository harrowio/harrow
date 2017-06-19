package stores

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type UnknownEventTypeError struct {
	EventName string
	RawData   string
}

func (store *UnknownEventTypeError) Error() string {
	return fmt.Sprintf("DbBillingEventStore: unknown event type %q", store.EventName)
}

type storedBillingEvent struct {
	*domain.BillingEvent
	RawData []byte `db:"raw_data"`
}

type DbBillingEventStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbBillingEventStore(tx *sqlx.Tx) *DbBillingEventStore {
	return &DbBillingEventStore{tx: tx}
}

func (store *DbBillingEventStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbBillingEventStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbBillingEventStore) Create(event *domain.BillingEvent) (uuid string, err error) {

	if !uuidhelper.IsValid(event.Uuid) {
		event.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO billing_events (uuid, organization_uuid, event_name, data) VALUES ($1, $2, $3, $4)`

	serializedEventData, err := json.Marshal(event.Data)
	if err != nil {
		return "", err
	}

	_, err = store.tx.Exec(q, event.Uuid, event.OrganizationUuid, event.EventName, serializedEventData)
	if err != nil {
		return "", err
	}

	return event.Uuid, nil
}

func (store *DbBillingEventStore) ReplayAllAfter(fn func(*domain.BillingEvent), after time.Time) error {
	q := `SELECT uuid, organization_uuid, event_name, occurred_on, data::text as raw_data FROM billing_events WHERE occurred_on >= $1`

	rows, err := store.tx.Queryx(q, after)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		storedEvent := &storedBillingEvent{}
		if err := rows.StructScan(&storedEvent); err != nil {
			return err
		}

		eventData := domain.NewBillingEventDataByName(storedEvent.EventName)
		if eventData == nil {
			return &UnknownEventTypeError{
				EventName: storedEvent.EventName,
				RawData:   string(storedEvent.RawData),
			}
		}

		if err := json.Unmarshal(storedEvent.RawData, eventData); err != nil {
			return err
		}

		storedEvent.BillingEvent.Data = eventData
		fn(storedEvent.BillingEvent)
	}

	return nil
}

func (store *DbBillingEventStore) FindAllByOrganizationUuid(organizationUuid string) ([]*domain.BillingEvent, error) {
	results := []*storedBillingEvent{}

	q := `SELECT uuid, organization_uuid, event_name, occurred_on, data::text as raw_data FROM billing_events WHERE organization_uuid = $1`

	err := store.tx.Select(&results, q, organizationUuid)
	if err != nil {
		return nil, err
	}

	deserialized := make([]*domain.BillingEvent, len(results))
	for i, result := range results {
		deserialized[i] = result.BillingEvent
		eventData := domain.NewBillingEventDataByName(result.EventName)
		if eventData == nil {
			return nil, &UnknownEventTypeError{
				EventName: result.EventName,
				RawData:   string(result.RawData),
			}
		}

		if err := json.Unmarshal(result.RawData, eventData); err != nil {
			return nil, err
		}

		deserialized[i].Data = eventData
	}

	return deserialized, nil
}
