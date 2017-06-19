package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
)

type DbScriptCardStore struct {
	tx DataSource
}

func NewDbScriptCardStore(tx DataSource) *DbScriptCardStore {
	return &DbScriptCardStore{
		tx: tx,
	}
}

func (self *DbScriptCardStore) FindAllByProjectUuid(projectUuid string) ([]*domain.ScriptCard, error) {
	q := `
SELECT
  tasks.project_uuid,
  tasks.uuid as script_uuid,
  tasks.name as script_name
FROM
  tasks
WHERE
  tasks.archived_at IS NULL
AND
  tasks.name NOT LIKE 'urn:harrow:%'
AND
  tasks.project_uuid = $1
;
`
	resolveErr := func(err error) ([]*domain.ScriptCard, error) {
		if err == sql.ErrNoRows {
			return []*domain.ScriptCard{}, nil
		}

		return nil, resolveErrType(err)
	}

	scripts := []*domain.ScriptCard{}
	if err := self.tx.Select(&scripts, q, projectUuid); err != nil {
		return resolveErr(err)
	}

	for _, script := range scripts {
		if err := self.findLastOperation(script); err != nil {
			return nil, err
		}
		if err := self.findEnabledEnvironments(script); err != nil {
			return nil, err
		}
		if err := self.findRecentOperationStatuses(script); err != nil {
			return nil, err
		}
	}

	return scripts, nil
}

func (self *DbScriptCardStore) findLastOperation(script *domain.ScriptCard) error {
	q := `
SELECT
  operations.*
FROM
  operations,
  jobs,
  tasks
WHERE
  jobs.task_uuid = tasks.uuid
AND
  operations.job_uuid = jobs.uuid
AND
  tasks.uuid = $1
AND
  operations.started_at IS NOT NULL
ORDER BY
  operations.started_at DESC
LIMIT 1
;`
	operation := &domain.Operation{}
	if err := self.tx.Get(operation, q, script.ScriptUuid); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	operation.GitLogs.Trim(5)
	script.LastOperation = operation

	return nil
}

func (self *DbScriptCardStore) findEnabledEnvironments(script *domain.ScriptCard) error {
	q := `
SELECT DISTINCT ON (environments.uuid)
  environments.*
FROM
  environments,
  jobs,
  tasks
WHERE
  jobs.task_uuid = tasks.uuid
AND
  jobs.environment_uuid = environments.uuid
AND
  tasks.uuid = $1
AND
  environments.archived_at IS NULL
AND
  environments.name NOT LIKE 'urn:harrow:%'
;`

	environments := []*domain.Environment{}
	if err := self.tx.Select(&environments, q, script.ScriptUuid); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	script.EnabledEnvironments = environments

	return nil
}

func (self *DbScriptCardStore) findRecentOperationStatuses(script *domain.ScriptCard) error {
	q := `
SELECT
  operations.*
FROM
  operations,
  jobs,
  tasks
WHERE
  jobs.task_uuid = tasks.uuid
AND
  operations.job_uuid = jobs.uuid
AND
  tasks.uuid = $1
AND
  jobs.environment_uuid = $2
ORDER BY operations.created_at DESC
LIMIT 5
;`
	script.RecentOperationStatusByEnvironmentUuid = map[string][]string{}
	for _, environment := range script.EnabledEnvironments {
		operations := []*domain.Operation{}
		if err := self.tx.Select(&operations, q, script.ScriptUuid, environment.Uuid); err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		}

		for _, operation := range operations {
			script.RecentOperationStatusByEnvironmentUuid[environment.Uuid] = append(
				script.RecentOperationStatusByEnvironmentUuid[environment.Uuid],
				operation.Status(),
			)
		}
	}

	return nil
}
