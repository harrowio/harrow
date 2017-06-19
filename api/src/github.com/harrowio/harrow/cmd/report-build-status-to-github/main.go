package reportBuildStatusToGitHub

import (
	"flag"
	"net/http"
	"net/url"
	"os"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/domain/interaction"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

const ProgramName = "report-build-status-to-github"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	operationUUID := flag.String("operation-uuid", "", "derive repositories and state from operation")
	repositoryUUID := flag.String("repository", "", "uuid of the repository to which to add the status")
	ref := flag.String("ref", "", "commit hash to which to add the status")
	state := flag.String("state", "success", "one of: success, error, pending, failure")
	targetURL := flag.String("target-url", "", "URL to show on GitHub")
	flag.Parse()

	conf := config.GetConfig()
	urls := conf.URLConfig()
	db, err := conf.DB()
	if err != nil {
		log.Fatal().Err(err)
	}

	if *operationUUID != "" {
		if err := runForOperation(log, urls.Base(), db, *operationUUID); err != nil {
			log.Fatal().Err(err)
		}
	} else {
		tx := db.MustBegin()
		defer tx.Rollback()
		datasource := NewDatasource(tx)
		report := interaction.NewReportBuildStatusToGitHub(
			*repositoryUUID,
			*ref,
			*state,
			*targetURL,
			log,
			http.DefaultClient,
			datasource,
			datasource,
		)

		if err := report.Execute(); err != nil {
			log.Fatal().Err(err)
		}
	}
}

func runForOperation(log logger.Logger, base *url.URL, db *sqlx.DB, operationUUID string) error {
	tx := db.MustBegin()
	defer tx.Rollback()
	operation, err := stores.NewDbOperationStore(tx).FindByUuid(operationUUID)
	if err != nil {
		return err
	}

	if operation.RepositoryCheckouts == nil {
		log.Info().Msgf("No repositories used by operation %s", operationUUID)
		return nil
	}
	datasource := NewDatasource(tx)
	state := "success"
	switch operation.Status() {
	case "canceled":
		state = "failure"
	case "timeout":
		state = "error"
	case "fatal":
		state = "error"
	case "failure":
		state = "failure"
	case "active":
		state = "pending"
	default:
		state = "success"
	}

	operationURL := base.ResolveReference(&url.URL{
		Path:     "/",
		Fragment: "/a/operations/" + operation.Uuid,
	}).String()

	for repository, checkouts := range operation.RepositoryCheckouts.Refs {
		for _, checkout := range checkouts {
			log.Info().Msgf("UPDATE repository %s %s %s", repository, checkout.Hash, checkout.Ref)
			report := interaction.NewReportBuildStatusToGitHub(repository, checkout.Hash, state, operationURL, log, http.DefaultClient, datasource, datasource)
			if err := report.Execute(); err != nil {
				log.Info().Msgf("repository %s: %s\n", repository, err)
			}
		}
	}

	return nil
}

type Datasource struct {
	tx *sqlx.Tx
}

func NewDatasource(tx *sqlx.Tx) *Datasource {
	return &Datasource{
		tx: tx,
	}
}

func (self *Datasource) FindRepository(repositoryUUID string) (*domain.Repository, error) {
	return stores.NewDbRepositoryStore(self.tx).FindByUuid(repositoryUUID)
}

func (self *Datasource) FindOAuthTokenForRepository(repositoryUUID string) (string, error) {
	token, err := stores.NewDbOAuthTokenStore(self.tx).FindByRepositoryUUID(repositoryUUID)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
