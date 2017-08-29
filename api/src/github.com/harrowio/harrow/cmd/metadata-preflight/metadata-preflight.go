package metadataPreflight

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	redis "gopkg.in/redis.v2"

	"golang.org/x/sys/unix"

	"github.com/boltdb/bolt"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/domain/interaction"
	"github.com/harrowio/harrow/git"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

type RepositoryStore interface {
	FindAllWithTriggers() ([]*domain.Repository, error)
	UpdateMetadata(repositoryUuid string, metadata *domain.RepositoryMetaData) error
	MarkAsAccessible(repositoryUuid string, accessible bool) error
}

type RepositoriesInDB struct {
	db *sqlx.DB
}

func (self *RepositoriesInDB) FindAllWithTriggers() ([]*domain.Repository, error) {
	tx := self.db.MustBegin()
	defer tx.Commit()
	store := stores.NewDbRepositoryStore(tx)
	return store.FindAllWithTriggers()
}

func (self *RepositoriesInDB) UpdateMetadata(repositoryUuid string, metadata *domain.RepositoryMetaData) error {
	tx := self.db.MustBegin()
	defer tx.Commit()
	store := stores.NewDbRepositoryStore(tx)
	return store.UpdateMetadata(repositoryUuid, metadata)
}

func (self *RepositoriesInDB) MarkAsAccessible(repositoryUuid string, accessible bool) error {
	tx := self.db.MustBegin()
	defer tx.Commit()
	store := stores.NewDbRepositoryStore(tx)
	return store.MarkAsAccessible(repositoryUuid, accessible)
}

const ProgramName = "metadata-preflight"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	clonedRepositories := map[string]*git.ClonedRepository{}

	dbFileName := flag.String("db", "/tmp/metadata-preflight-cache.db", "Cache db file for ls-remote output")
	pollingInterval := flag.Duration("poll", 1*time.Minute, "Interval at which to scan for differences in ls-remote hash.")
	runOnce := flag.Bool("once", false, "run once and then exit")

	flag.Parse()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, unix.SIGTERM)

	interval := time.Tick(*pollingInterval)

	run(config.GetConfig(), *dbFileName, clonedRepositories)
	if *runOnce {
		fmt.Fprintf(os.Stderr, "Exiting after one run, was given -once\n")
		os.Exit(0)
	}

	for {
		select {
		case <-interval:
			run(config.GetConfig(), *dbFileName, clonedRepositories)
		case s := <-signals:
			log.Info().Msgf("Got signal: %#v", s)
			os.Exit(0)
		}
	}

}

func run(c *config.Config, dbFileName string, clonedRepositories map[string]*git.ClonedRepository) {
	db, err := c.DB()
	if err != nil {
		log.Fatal().Err(err)
	}
	defer db.Close()
	bDb, err := bolt.Open(dbFileName, 0644, nil)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer bDb.Close()

	activitySink := activity.NewAMQPTransport(c.AmqpConnectionString(), "metadata-preflight-worker")
	defer activitySink.Close()

	repoStore := &RepositoriesInDB{db: db}

	repos, err := repoStore.FindAllWithTriggers()
	if err != nil {
		log.Fatal().Msgf("Error querying repositories handle: %s", err)
	}

	for _, repository := range repos {
		log.Info().Str("url", repository.Url).Str("uuid", repository.Uuid).Msgf("updating metadata for repository")
		start := time.Now()
		UpdateMetadata(c, db, repository, activitySink, repoStore, clonedRepositories)
		elapsed := time.Since(start)
		log.Info().Str("url", repository.Url).Str("uuid", repository.Uuid).Int64("repo-meta-update-ns", elapsed.Nanoseconds()).Msg("finished")
		log.Info().Str("url", repository.Url).Str("uuid", repository.Uuid).Int64("repo-meta-update-ms", elapsed.Nanoseconds()/int64(1e6)).Msg("finished")
	}
}

func UpdateMetadata(c *config.Config, db *sqlx.DB, repository *domain.Repository, activitySink activity.Sink, repositoryStore RepositoryStore, clonedRepositories map[string]*git.ClonedRepository) {
	clonedRepository := clonedRepositoryFor(c, db, repository, clonedRepositories)
	if clonedRepository == nil {
		return
	}

	if err := clonedRepository.Pull(); err != nil {
		log.Error().Msgf("clonedrepository.pull(): %s", err)
		return
	}

	metadata, err := clonedRepository.FetchMetadata()
	if err != nil {
		log.Error().Msgf("clonedrepository.fetchmetadata(): %s", err)
		return
	}

	repositoryMetadata := domain.NewRepositoryMetaData()
	for _, contributor := range metadata.Contributors {
		repositoryMetadata.Contributors[contributor.Email] = &domain.Person{
			Name:  contributor.Name,
			Email: contributor.Email,
		}
	}

	for ref, hash := range metadata.Refs {
		repositoryMetadata.Refs[ref] = hash
	}

	update := interaction.NewUpdateRepositoryMetaData(interaction.NewBusActivitySink(activitySink, log), repositoryStore)
	if err := update.Update(repository.Uuid, repository.Metadata, repositoryMetadata); err != nil {
		log.Error().Msgf("updaterepositorymetadata: %s: %s", repository.Uuid, err)
		return
	}

	repositoryStore.MarkAsAccessible(repository.Uuid, true)
}

func clonedRepositoryFor(c *config.Config, db *sqlx.DB, repository *domain.Repository, clonedRepositories map[string]*git.ClonedRepository) *git.ClonedRepository {
	existing, found := clonedRepositories[repository.Uuid]
	if found {
		return existing
	}

	tx := db.MustBegin()
	defer tx.Rollback()
	redisClient := redis.NewTCPClient(c.RedisConnOpts(1))
	secrets := stores.NewRedisSecretKeyValueStore(redisClient)
	repositoryCredentials := stores.NewRepositoryCredentialStore(secrets, tx)
	OS := git.NewOperatingSystem(c.FilesystemConfig().GitTempDir)
	clonedRepository, err := repository.ClonedGit(OS, repositoryCredentials)
	if err != nil {
		log.Error().Msgf("repository(%q).clonedgit: %s", repository.Uuid, err)
		return nil
	}

	clonedRepositories[repository.Uuid] = clonedRepository

	return clonedRepository
}
