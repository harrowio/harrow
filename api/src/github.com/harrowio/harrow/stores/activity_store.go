package stores

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
)

type DbActivityStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbActivityStore(tx *sqlx.Tx) *DbActivityStore {
	return &DbActivityStore{tx: tx}
}

func (self *DbActivityStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbActivityStore) SetLogger(l logger.Logger) {
	self.log = l
}

type activity struct {
	Id              int            `db:"id"`
	Name            string         `db:"name"`
	OccurredOn      time.Time      `db:"occurred_on"`
	CreatedAt       time.Time      `db:"created_at"`
	ContextUserUuid *string        `db:"context_user_uuid"`
	Payload         types.JSONText `db:"payload"`
	Extra           types.JSONText `db:"extra"`
}

func (self *activity) toActivity() (*domain.Activity, error) {
	result := &domain.Activity{
		Id:              self.Id,
		Name:            self.Name,
		OccurredOn:      self.OccurredOn,
		CreatedAt:       self.CreatedAt,
		ContextUserUuid: self.ContextUserUuid,
	}

	if err := self.Extra.Unmarshal(&result.Extra); err != nil {
		return nil, err
	}

	if err := activities.UnmarshalPayload(result, []byte(self.Payload)); err != nil {
		return nil, err
	}

	return result, nil
}

func bindActivity(from *domain.Activity) (*activity, error) {
	result := &activity{
		Name:            from.Name,
		OccurredOn:      from.OccurredOn,
		ContextUserUuid: from.ContextUserUuid,
	}

	payload, err := json.Marshal(from.Payload)
	if err != nil {
		return result, err
	}
	result.Payload = types.JSONText(payload)

	extra, err := json.Marshal(from.Extra)
	if err != nil {
		return result, err
	}
	result.Extra = types.JSONText(extra)

	return result, nil
}

func (self *DbActivityStore) Store(activity *domain.Activity) error {

	q := `INSERT INTO activities (name, occurred_on, context_user_uuid, payload, extra) VALUES (:name, :occurred_on, :context_user_uuid, :payload, :extra) RETURNING id`

	id := 0
	params, err := bindActivity(activity)
	if err != nil {
		return resolveErrType(err)
	}
	query, args, err := self.tx.BindNamed(q, params)

	// fmt.Fprintf(os.Stderr, "%s %v\n", query, args)
	if err != nil {
		return resolveErrType(err)
	}
	if err := self.tx.Get(&id, query, args...); err != nil {
		return resolveErrType(err)
	}
	activity.Id = id

	return nil
}

func (self *DbActivityStore) AllByNameSince(activityNames []string, since time.Time, handler func(*domain.Activity) error) error {

	inList := bytes.NewBufferString("")

	// Build order by clause to sort by specific activity names.
	// Example:
	//
	//    ORDER BY (name = 'job.added') DESC,
	//             (name = 'job.edited') DESC
	//
	orderByClause := bytes.NewBufferString(`ORDER BY `)
	for i, name := range activityNames {
		fmt.Fprintf(orderByClause, `(name = '%s') DESC`, name)
		fmt.Fprintf(inList, `'%s'`, name)
		if i < len(activityNames)-1 {
			fmt.Fprintf(orderByClause, `, `)
			fmt.Fprintf(inList, `, `)
		}
	}
	fmt.Fprintf(orderByClause, `, occurred_on ASC`)
	q := fmt.Sprintf(`SELECT * FROM activities WHERE occurred_on > $1 AND name IN (%s) %s`, inList, orderByClause)

	rows, err := self.tx.Queryx(q, since)
	if err != nil {
		return resolveErrType(err)
	}
	defer rows.Close()

	for rows.Next() {
		row := activity{}

		if err := rows.StructScan(&row); err != nil {
			return resolveErrType(err)
		}
		activity, err := row.toActivity()
		if err != nil {
			return resolveErrType(err)
		}

		if err := handler(activity); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (self *DbActivityStore) AllByUser(userUuid string, handler func(*domain.Activity) error) error {

	q := `SELECT * FROM activities WHERE context_user_uuid = $1 ORDER BY occurred_on ASC`
	rows, err := self.tx.Queryx(q, userUuid)
	if err != nil {
		return resolveErrType(err)
	}
	defer rows.Close()

	for rows.Next() {
		row := activity{}

		if err := rows.StructScan(&row); err != nil {
			return resolveErrType(err)
		}
		activity, err := row.toActivity()
		if err != nil {
			return resolveErrType(err)
		}

		if err := handler(activity); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (self *DbActivityStore) AllByProjectSince(projectUuid string, since time.Time, handler func(*domain.Activity) error) error {

	q := `SELECT * FROM activities WHERE (extra->>'projectUuid') = $1 AND occurred_on > $2 ORDER BY occurred_on ASC`
	rows, err := self.tx.Queryx(q, projectUuid, since)
	if err != nil {
		return resolveErrType(err)
	}
	defer rows.Close()

	for rows.Next() {
		row := activity{}

		if err := rows.StructScan(&row); err != nil {
			return resolveErrType(err)
		}
		activity, err := row.toActivity()
		if err != nil {
			return resolveErrType(err)
		}

		if err := handler(activity); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (self *DbActivityStore) All(handler func(*domain.Activity) error) error {
	return self.AllSince(time.Time{}, handler)
}

func (self *DbActivityStore) AllSince(since time.Time, handler func(*domain.Activity) error) error {

	q := `SELECT * FROM activities WHERE occurred_on > $1 ORDER BY occurred_on ASC`
	rows, err := self.tx.Queryx(q, since)
	if err != nil {
		return resolveErrType(err)
	}
	defer rows.Close()

	for rows.Next() {
		row := activity{}

		if err := rows.StructScan(&row); err != nil {
			return resolveErrType(err)
		}
		activity, err := row.toActivity()
		if err != nil {
			return resolveErrType(err)
		}

		if err := handler(activity); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (self *DbActivityStore) FindActivityByNameAndPayloadUuid(name string, payloadUuid string) (*domain.Activity, error) {
	jsonExpr := `(payload->'uuid')::text`
	if name == "organization.created" {
		jsonExpr = `(payload->'Organization'->'uuid')::text`
	}

	q := fmt.Sprintf(`SELECT * FROM activities WHERE name = $1 AND trim(both '"' from %s) = $2 LIMIT 1`, jsonExpr)
	result := activity{}
	err := self.tx.Get(&result, q, name, payloadUuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return result.toActivity()
}

func (self *DbActivityStore) FindActivityById(id int) (*domain.Activity, error) {

	q := `SELECT * FROM activities WHERE id = $1`
	result := activity{}
	err := self.tx.Get(&result, q, id)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return result.toActivity()
}
