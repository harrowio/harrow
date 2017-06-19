package harrowUpdateRepositoryMetadata

import (
	"fmt"
	"os"
	"path/filepath"

	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/domain/interaction"
	"github.com/harrowio/harrow/git"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/bus/activity"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

const ProgramName = "harrow-update-repository-metadata"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	repositoryUUID := os.Args[1]

	conf := config.GetConfig()
	db, err := conf.DB()
	if err != nil {
		log.Fatal().Msgf("config.GetConfig().DB(): %s", err)
	}

	tx := db.MustBegin()
	defer tx.Rollback()
	repositoryStore := stores.NewDbRepositoryStore(tx)
	repository, err := repositoryStore.FindByUuid(repositoryUUID)
	if err != nil {
		log.Fatal().Msgf("stores.NewDbRepositoryStore(tx).FindByUuid(%q): %s", repositoryUUID, err)
	}
	redisClient := redis.NewTCPClient(conf.RedisConnOpts(1))
	defer redisClient.Close()

	secretStore := stores.NewRedisSecretKeyValueStore(redisClient)
	repoCredentialStore := stores.NewRepositoryCredentialStore(secretStore, tx)

	activitySink := activity.NewAMQPTransport(
		conf.AmqpConnectionString(),
		fmt.Sprintf("%s-%d", filepath.Base(os.Args[0]), os.Getpid()),
	)

	OS := git.NewOperatingSystem(conf.FilesystemConfig().GitTempDir)
	clonedRepository, err := repository.ClonedGit(OS, repoCredentialStore)
	if err != nil {
		log.Fatal().Msgf("repository(%q).ClonedGit(): %s", repository.Uuid, err)
	}
	tx.Rollback()

	metadata, err := clonedRepository.FetchMetadata()
	if err != nil {
		log.Fatal().Msgf("clonedRepository.FetchMetadata(): %s", err)
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

	// new transaction to avoid long-running transactions, because
	// cloning a repository can take longer than the configured
	// transaction timeout
	shortTx := db.MustBegin()
	defer shortTx.Commit()
	repositoryStore = stores.NewDbRepositoryStore(shortTx)

	update := interaction.NewUpdateRepositoryMetaData(interaction.NewBusActivitySink(activitySink, log), repositoryStore)
	if err := update.Update(repository.Uuid, repository.Metadata, repositoryMetadata); err != nil {
		log.Fatal().Msgf("UpdateRepositoryMetadata: %s: %s", repository.Uuid, err)
	}

	repositoryStore.MarkAsAccessible(repository.Uuid, true)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:

%s <repository-uuid>
`, os.Args[0])
}
