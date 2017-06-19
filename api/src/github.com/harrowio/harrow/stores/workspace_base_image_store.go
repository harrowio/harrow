package stores

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type DbWorkspaceBaseImageStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbWorkspaceBaseImageStore(tx *sqlx.Tx) *DbWorkspaceBaseImageStore {
	return &DbWorkspaceBaseImageStore{tx: tx}
}

func (store *DbWorkspaceBaseImageStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbWorkspaceBaseImageStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbWorkspaceBaseImageStore) Create(wsbi *domain.WorkspaceBaseImage) (string, error) {

	q := `INSERT INTO workspace_base_images (name, repository, path, ref, type)
	VALUES (:name, :repository, :path, :ref, :type) RETURNING uuid`

	rows, err := store.tx.NamedQuery(q, wsbi)
	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&wsbi.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return wsbi.Uuid, nil
}

func (store *DbWorkspaceBaseImageStore) FindByUuid(wsbiUuid string) (*domain.WorkspaceBaseImage, error) {

	result := domain.WorkspaceBaseImage{}
	q := `SELECT * FROM workspace_base_images WHERE uuid = $1;`
	if err := store.tx.Get(&result, q, wsbiUuid); err != nil {
		return nil, resolveErrType(err)
	} else {
		return &result, nil
	}
}
