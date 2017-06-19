package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/git"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

func ReadRepoParams(r io.Reader) (*repoParams, error) {
	decoder := json.NewDecoder(r)
	var w repoParamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

type repoParamsWrapper struct {
	Subject repoParams
}

type repoParams struct {
	Uuid           string `json:"uuid"`
	Url            string `json:"url"`
	Name           string `json:"name"`
	GithubImported bool   `json:"githubImported"`
	GithubLogin    string `json:"githubLogin"`
	GithubRepo     string `json:"githubRepo"`
	ProjectUuid    string `json:"projectUuid"`
}

func copyRepoParams(p *repoParams, m *domain.Repository) {
	m.Uuid = p.Uuid
	m.ProjectUuid = p.ProjectUuid
	m.Url = p.Url
	m.Name = p.Name
	m.GithubImported = p.GithubImported
	m.GithubLogin = p.GithubLogin
	m.GithubRepo = p.GithubRepo
}

func MountRepoHandler(r *mux.Router, ctxt ServerContext) {

	rh := repoHandler{
		ss: ctxt.SecretKeyValueStore(),
	}

	// Collection
	root := r.PathPrefix("/repositories").Subrouter()

	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Methods("GET").Path("/operations").Handler(HandlerFunc(ctxt, rh.Operations)).
		Name("repository-operations")
	related.Methods("POST").Path("/checks").Queries("updateMetadata", "yes").Handler(HandlerFunc(ctxt, rh.MetaData)).
		Name("repositories-check-metadata")

	related.Methods("POST").Path("/checks").Handler(HandlerFunc(ctxt, rh.Checks)).
		Name("repository-checks")
	related.Methods("POST").Path("/metadata").Handler(HandlerFunc(ctxt, rh.MetaData)).
		Name("repository-metadata")

	related.Methods("GET").Path("/credential").Handler(HandlerFunc(ctxt, rh.Credential)).
		Name("repository-credential")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, rh.Show)).
		Name("repository-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, rh.Archive)).
		Name("repository-archive")

	root.Methods("POST").Handler(HandlerFunc(ctxt, rh.Create)).
		Name("repository-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, rh.Update)).
		Name("repository-update")
}

type repoHandler struct {
	ss stores.SecretKeyValueStore
}

func (self repoHandler) Show(ctxt RequestContext) (err error) {

	repoStore := stores.NewDbRepositoryStore(ctxt.Tx())

	repo, err := repoStore.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(repo); !allowed {
		return err
	}

	writeAsJson(ctxt, repo)

	return err

}

func (self repoHandler) Create(ctxt RequestContext) error {

	params, err := ReadRepoParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	store := stores.NewDbRepositoryStore(ctxt.Tx())
	repo := &domain.Repository{}
	copyRepoParams(params, repo)

	if err := domain.ValidateRepository(repo); err != nil {
		return err
	}

	gitRepo, err := git.NewRepository(repo.Url)
	if err != nil {
		return domain.NewValidationError("url", err.Error())
	}

	userInfo := gitRepo.URL.User
	if gitRepo.URL.User != nil {
		gitRepo.URL.User = nil
		repo.Url = gitRepo.URL.String()
	}

	if allowed, err := ctxt.Auth().CanCreate(repo); !allowed {
		return err
	}

	if allowed, err := limits.CanCreate(repo); !allowed {
		return err
	}

	uuid, err := store.Create(repo)
	if err != nil {
		return err
	}

	if _, err := makeSshRepositoryCredential(ctxt, uuid); err != nil {
		return err
	}

	if gitRepo.UsesHTTP() {
		_, err := makeHTTPSRepositoryCredential(ctxt, userInfo, uuid)
		if err != nil {
			return err
		}
	}

	ctxt.EnqueueActivity(activities.RepositoryAdded(repo), nil)

	repo, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if gitRepo.IsAccessible() {
		if (gitRepo.UsesHTTP() && userInfo == nil) || gitRepo.UsesSSH() {
			ctxt.EnqueueActivity(activities.RepositoryDetectedAsPublic(repo), nil)
		}
		ctxt.EnqueueActivity(activities.RepositoryConnectedSuccessfully(repo), nil)
	} else {
		ctxt.EnqueueActivity(activities.RepositoryDetectedAsPrivate(repo), nil)
	}

	ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), repo))
	ctxt.W().Header().Add("Content-Type", "application/json")
	writeAsJson(ctxt, repo)

	return nil
}

func (self repoHandler) Update(ctxt RequestContext) (err error) {

	params, err := ReadRepoParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	store := stores.NewDbRepositoryStore(ctxt.Tx())

	var repo *domain.Repository

	repo, err = store.FindByUuid(params.Uuid)
	if err != nil {
		return err
	}

	copyRepoParams(params, repo)

	if err := domain.ValidateRepository(repo); err != nil {
		return err
	}

	gitRepo, err := git.NewRepository(repo.Url)
	if err != nil {
		return domain.NewValidationError("url", err.Error())
	}
	userInfo := gitRepo.URL.User
	if gitRepo.UsesHTTP() {
		gitRepo.URL.User = nil
		repo.Url = gitRepo.URL.String()
	}

	var uuid string

	if allowed, err := ctxt.Auth().CanUpdate(repo); !allowed {
		return err
	}

	if err := store.Update(repo); err != nil {
		return err
	}

	uuid = repo.Uuid
	ctxt.EnqueueActivity(activities.RepositoryEdited(repo), nil)

	if userInfo != nil && gitRepo.UsesHTTP() {
		toSave := (*domain.BasicRepositoryCredential)(nil)
		repositoryCredentials := stores.NewRepositoryCredentialStore(ctxt.SecretKeyValueStore(), ctxt.Tx())
		existingCredential, err := repositoryCredentials.FindByRepositoryUuidAndType(repo.Uuid, domain.RepositoryCredentialBasic)
		if err != nil && !domain.IsNotFound(err) {
			return err
		}

		password, _ := userInfo.Password()
		if existingCredential != nil {
			toSave = &domain.BasicRepositoryCredential{
				RepositoryCredential: existingCredential,
				Username:             userInfo.Username(),
				Password:             password,
			}
		} else {
			toSave = &domain.BasicRepositoryCredential{
				RepositoryCredential: &domain.RepositoryCredential{
					Uuid:           uuidhelper.MustNewV4(),
					Name:           "HTTP Access",
					RepositoryUuid: repo.Uuid,
					Type:           domain.RepositoryCredentialBasic,
					Status:         domain.RepositoryCredentialPresent,
				},
				Username: userInfo.Username(),
				Password: password,
			}
		}

		credential, err := toSave.AsRepositoryCredential()
		if err != nil {
			return err
		}
		if existingCredential == nil {
			if _, err := repositoryCredentials.Create(credential); err != nil {
				return err
			}
		} else {
			if err := repositoryCredentials.Update(credential); err != nil {
				return err
			}
		}
	}

	repo, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if gitRepo.IsAccessible() {
		if (gitRepo.UsesHTTP() && userInfo == nil) || gitRepo.UsesSSH() {
			ctxt.EnqueueActivity(activities.RepositoryDetectedAsPublic(repo), nil)
		}
		if repo.ConnectedSuccessfully != nil && !*repo.ConnectedSuccessfully {
			ctxt.EnqueueActivity(activities.RepositoryConnectedSuccessfully(repo), nil)
		}
	} else {
		ctxt.EnqueueActivity(activities.RepositoryDetectedAsPrivate(repo), nil)
	}

	writeAsJson(ctxt, repo)

	return err

}

func (self repoHandler) Archive(ctxt RequestContext) (err error) {

	repoUuid := ctxt.PathParameter("uuid")

	store := stores.NewDbRepositoryStore(ctxt.Tx())
	repo, err := store.FindByUuid(repoUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(repo); !allowed {
		return err
	}

	err = store.ArchiveByUuid(repoUuid)
	if err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusNoContent)

	return err
}

func (self repoHandler) Checks(ctxt RequestContext) (err error) {

	repoUuid := ctxt.PathParameter("uuid")
	repoStore := stores.NewDbRepositoryStore(ctxt.Tx())

	repo, err := repoStore.FindByUuid(repoUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(repo); !allowed {
		return err
	}

	repositoryCredentials := stores.NewRepositoryCredentialStore(ctxt.SecretKeyValueStore(), ctxt.Tx())
	gitRepo, err := repo.ClonedGit(OS, repositoryCredentials)
	if err != nil {
		return NewInternalError(err)
	}

	accessible, err := gitRepo.IsAccessible()
	if err != nil {
		ctxt.Log().Info().Msgf("gitrepo(%q).isaccessible: %s", repo.Uuid, err)
	}
	if err := repoStore.MarkAsAccessible(repoUuid, accessible); err != nil {
		return err
	}

	if accessible && (repo.ConnectedSuccessfully != nil && !*repo.ConnectedSuccessfully) {
		if repo.ConnectedSuccessfully != nil && !*repo.ConnectedSuccessfully {
			ctxt.EnqueueActivity(activities.RepositoryConnectedSuccessfully(repo), nil)
		}
	}

	ctxt.W().Header().Set("Content-Type", "application/json")
	if accessible {
		fmt.Fprintf(ctxt.W(), `{"accessible":true}`)
	} else {
		fmt.Fprintf(ctxt.W(), `{"accessible":false}`)
	}

	return nil
}

func (self repoHandler) MetaData(ctxt RequestContext) (err error) {

	repoUuid := ctxt.PathParameter("uuid")
	repoStore := stores.NewDbRepositoryStore(ctxt.Tx())

	repo, err := repoStore.FindByUuid(repoUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(repo); !allowed {
		return err
	}

	opStore := stores.NewDbOperationStore(ctxt.Tx())

	op := repo.NewMetadataUpdateOperation()

	op.Parameters.Reason = domain.OperationTriggeredByUser
	op.Uuid, err = opStore.Create(op)
	if err != nil {
		ctxt.Log().Error().Msgf("error creating operation: %s\n", err)
	}

	writeAsJson(ctxt, op)

	return err

}

func (self repoHandler) Operations(ctxt RequestContext) (err error) {

	store := stores.NewDbRepositoryStore(ctxt.Tx())
	operationStore := stores.NewDbOperationStore(ctxt.Tx())

	repo, err := store.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(repo); !allowed {
		return err
	}

	operations, err := operationStore.FindAllByRepositoryUuid(repo.Uuid)
	if err != nil {
		return
	}

	var interfaceOperations []interface{} = make([]interface{}, 0, len(operations))
	for _, o := range operations {
		if allowed, _ := ctxt.Auth().CanRead(o); allowed {
			interfaceOperations = append(interfaceOperations, o)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceOperations),
		Count:      len(interfaceOperations),
		Collection: interfaceOperations,
	})

	return err

}

func (self repoHandler) Credential(ctxt RequestContext) (err error) {

	store := stores.NewRepositoryCredentialStore(self.ss, ctxt.Tx())
	repoUuid := ctxt.PathParameter("uuid")
	repository, err := stores.NewDbRepositoryStore(ctxt.Tx()).FindByUuid(repoUuid)
	if err != nil {
		return err
	}

	credentialType := domain.RepositoryCredentialType(ctxt.R().URL.Query().Get("type"))
	if credentialType == "" {
		if strings.HasPrefix(repository.Url, "http") {
			credentialType = domain.RepositoryCredentialBasic
		} else {
			credentialType = domain.RepositoryCredentialSsh
		}
	}

	rc, err := store.FindByRepositoryUuidAndType(repoUuid, credentialType)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(rc); !allowed {
		return err
	}

	if rc.IsSsh() {
		if rc.IsPending() {
			writeAsJson(ctxt, rc)
			return nil
		}

		credentialWithVisiblePk := &struct {
			*domain.RepositoryCredential
			PublicKey string `json:"publicKey"`
		}{RepositoryCredential: rc}

		err = json.Unmarshal(rc.SecretBytes, credentialWithVisiblePk)
		if err != nil {
			return err
		}

		writeAsJson(ctxt, credentialWithVisiblePk)
		return nil
	} else if rc.IsBasic() {
		basicCredential, err := domain.AsBasicRepositoryCredential(rc)
		if err != nil {
			return err
		}
		result := struct {
			*domain.RepositoryCredential
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			RepositoryCredential: rc,
			Username:             basicCredential.Username,
			Password:             basicCredential.Password,
		}
		writeAsJson(ctxt, &result)
		return nil
	}
	return fmt.Errorf("rendering non-ssh RepositoryCredentials not implemented")
}

func makeSshRepositoryCredential(ctxt RequestContext, repositoryUuid string) (*domain.RepositoryCredential, error) {
	tx := ctxt.Tx()
	keyName := fmt.Sprintf("repository-%s@harrow.io", repositoryUuid)
	rc := &domain.RepositoryCredential{}
	rc.Type = domain.RepositoryCredentialSsh
	rc.Name = keyName
	rc.RepositoryUuid = repositoryUuid

	store := stores.NewRepositoryCredentialStore(ctxt.SecretKeyValueStore(), tx)
	uuid, err := store.Create(rc)
	if err != nil {
		return nil, err
	}
	reloaded, err := store.FindByUuid(uuid)
	if err != nil {
		return nil, err
	}

	return reloaded, nil
}

func makeHTTPSRepositoryCredential(ctxt RequestContext, userInfo *url.Userinfo, repositoryUuid string) (*domain.RepositoryCredential, error) {
	if userInfo == nil {
		return nil, nil
	}

	repositoryCredentials := stores.NewRepositoryCredentialStore(ctxt.SecretKeyValueStore(), ctxt.Tx())

	password, _ := userInfo.Password()
	toSave := &domain.BasicRepositoryCredential{
		RepositoryCredential: &domain.RepositoryCredential{
			Uuid:           uuidhelper.MustNewV4(),
			Name:           "HTTP Access",
			RepositoryUuid: repositoryUuid,
			Type:           domain.RepositoryCredentialBasic,
			Status:         domain.RepositoryCredentialPresent,
		},
		Username: userInfo.Username(),
		Password: password,
	}

	credential, err := toSave.AsRepositoryCredential()
	if err != nil {
		return nil, err
	}

	credentialUuid, err := repositoryCredentials.Create(credential)
	if err != nil {
		return nil, err
	}

	reloaded, err := repositoryCredentials.FindByUuid(credentialUuid)
	if err != nil {
		return nil, err
	}

	return reloaded, nil
}
