package keymaker

import (
	"os"
	"time"

	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/sshkey"
	"github.com/harrowio/harrow/stores"
	"github.com/rs/zerolog"

	"encoding/json"

	"github.com/jmoiron/sqlx"
)

type createEvent struct {
	New struct {
		Uuid string
	}
}

const ProgramName = "keymaker"

type keymaker struct {
	logger      logger.Logger
	config      *config.Config
	redisClient *redis.Client
}

func (km *keymaker) generateSecret(c *config.Config, redisClient *redis.Client, message broadcast.Message) {
	db, err := km.config.DB()
	if err != nil {
		message.RequeueAfter(10 * time.Second)
		km.logger.Warn().Err(err)
		return
	}

	tx, err := db.Beginx()
	if err != nil {
		km.logger.Info().Msgf("Unable to begin tx: %s\n", err)
		message.RequeueAfter(10 * time.Second)
		return
	}
	defer tx.Rollback()

	secretUuid := message.UUID()
	ss := stores.NewRedisSecretKeyValueStore(redisClient)
	store := stores.NewSecretStore(ss, tx)
	store.SetLogger(km.logger)

	secret, err := store.FindByUuid(secretUuid)
	if err != nil {
		km.logger.Info().Msgf("Could not find indicated secret %s: %s\n", secretUuid, err)
		message.RejectForever()
		return
	}
	if !secret.IsSsh() || !secret.IsPending() {
		km.logger.Debug().Msgf("Not acting on %#v\n", secret)
		message.Acknowledge()
		return
	}
	err = km.generateSecretBytes(secret, tx)
	if err != nil {
		km.logger.Info().Msgf("Error generating key bytes for %s: %s\n", secretUuid, err)
		message.RequeueAfter(10 * time.Second)
		return
	}
	km.logger.Info().Str("Generated key for Secret", secretUuid)
	if err := tx.Commit(); err != nil {
		km.logger.Warn().Msgf("generateSecret: tx.Commit: %s", err)
		message.RequeueAfter(10 * time.Second)
	} else {
		message.Acknowledge()
	}
}

func (km *keymaker) generateRepositoryCredential(activityBus activity.Sink, message broadcast.Message) {
	db, err := km.config.DB()
	if err != nil {
		message.RequeueAfter(10 * time.Second)
		km.logger.Warn().Err(err)
		return
	}

	repositoryUuid := message.UUID()
	tx := db.MustBegin()
	defer tx.Rollback()
	ss := stores.NewRedisSecretKeyValueStore(km.redisClient)
	repositoryCredentialsStore := stores.NewRepositoryCredentialStore(ss, tx)
	rc, err := repositoryCredentialsStore.FindByRepositoryUuidAndType(repositoryUuid, domain.RepositoryCredentialSsh)
	if err != nil {
		km.logger.Error().Msgf("SSH credential not found for %q: %s\n", repositoryUuid, err)
		message.RejectForever()
		return
	}

	km.logger.Info().Str("Made new RepositoryCredential for", repositoryUuid)
	err = km.generateRepositoryCredentialBytes(db, rc)
	if err != nil {
		km.logger.Info().Msgf("Error generating secret bytes for SshRepositoryCredential %s: %s\n", repositoryUuid, err)
		message.RejectForever()
		return
	}
	if err := activityBus.Publish(activities.DeployKeyGenerated(rc)); err != nil {
		km.logger.Info().Msgf("Error emitting activity: %s", err)
	}
	km.logger.Info().Str("Generated key bytes for", repositoryUuid)
	message.Acknowledge()
}

func mustParseUuid(body []byte) string {
	var create createEvent
	err := json.Unmarshal(body, &create)
	if err != nil {
		panic(err)
	}
	return create.New.Uuid
}

func (km *keymaker) generateSecretBytes(secret *domain.Secret, tx *sqlx.Tx) error {
	ss := stores.NewRedisSecretKeyValueStore(km.redisClient)
	store := stores.NewSecretStore(ss, tx)
	store.SetLogger(km.logger)

	priv, pub, err := sshkey.Generate(secret.Name, "rsa", 8192)
	if err != nil {
		return err
	}
	secret.Status = domain.SecretPresent
	sshSecret := domain.SshSecret{Secret: secret}
	sshSecret.PrivateKey = priv
	sshSecret.PublicKey = pub
	secret, err = sshSecret.AsSecret()
	if err != nil {
		return err
	}
	return store.Update(secret)
}

func (km *keymaker) generateRepositoryCredentialBytes(db *sqlx.DB, repositoryCredential *domain.RepositoryCredential) error {
	tx, err := db.Beginx()
	if err != nil {
		km.logger.Info().Msgf("Unable to begin tx: %s\n", err)
		return err
	}
	defer tx.Rollback()
	ss := stores.NewRedisSecretKeyValueStore(km.redisClient)
	store := stores.NewRepositoryCredentialStore(ss, tx)

	priv, pub, err := sshkey.Generate(repositoryCredential.Name, "rsa", 8192)
	if err != nil {
		return err
	}
	repositoryCredential.Status = domain.RepositoryCredentialPresent
	sshRc := domain.SshRepositoryCredential{RepositoryCredential: repositoryCredential}
	sshRc.PrivateKey = priv
	sshRc.PublicKey = pub
	repositoryCredential, err = sshRc.AsRepositoryCredential()
	if err != nil {
		return err
	}
	err = store.Update(repositoryCredential)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (km *keymaker) rekey() error {
	db, err := km.config.DB()
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	repositories, err := stores.NewDbRepositoryStore(tx).FindAll()
	if err != nil {
		return err
	}
	ss := stores.NewRedisSecretKeyValueStore(km.redisClient)
	store := stores.NewSecretStore(ss, tx)
	store.SetLogger(km.logger)

	secrets, err := store.FindAll()
	repositoryCredentialsStore := stores.NewRepositoryCredentialStore(ss, tx)
	if err != nil {
		return err
	}
	for _, repository := range repositories {
		rc, err := repositoryCredentialsStore.FindByRepositoryUuid(repository.Uuid)
		err = km.generateRepositoryCredentialBytes(db, rc)
		if err != nil {
			return err
		}
		km.logger.Info().Str("Rekeyed Repository", repository.Uuid)
	}
	for _, secret := range secrets {
		if secret.IsSsh() {
			err := km.generateSecretBytes(secret, tx)
			if err != nil {
				return err
			}
			km.logger.Info().Str("Rekeyed Secret", secret.Uuid)
		}
	}

	tx.Commit()
	return nil
}

func (km *keymaker) rekeyMissing() error {
	db, err := km.config.DB()
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	repositories, err := stores.NewDbRepositoryStore(tx).FindAll()
	if err != nil {
		return err
	}
	ss := stores.NewRedisSecretKeyValueStore(km.redisClient)
	secretStore := stores.NewSecretStore(ss, tx)
	secretStore.SetLogger(km.logger)

	repositoryCredentialsStore := stores.NewRepositoryCredentialStore(ss, tx)

	secrets, err := secretStore.FindAll()
	if err != nil {
		return err
	}
	km.logger.Info().Msgf("scanning %d repositories\n", len(repositories))
	for _, repository := range repositories {
		rc, err := repositoryCredentialsStore.FindByRepositoryUuidAndTypeNoLoad(repository.Uuid, domain.RepositoryCredentialSsh)

		if err != nil && domain.IsNotFound(err) {
			km.logger.Info().Str("skipped repository", repository.Uuid)
			continue
		}

		if _, err := ss.Get(rc.Uuid, rc.Key); err == stores.ErrKeyNotFound {
			km.logger.Info().Str("generated for repository", repository.Uuid)
			err = km.generateRepositoryCredentialBytes(db, rc)
			if err != nil {
				km.logger.Info().Str("erred on repository", repository.Uuid).Err(err)
			} else {
				km.logger.Info().Str("finished repository", repository.Uuid)
			}
		}
	}

	km.logger.Info().Msgf("scanning %d secrets", len(secrets))
	for _, secret := range secrets {
		if secret.IsSsh() {
			if _, err := ss.Get(secret.Uuid, secret.Key); err != nil {
				if err == stores.ErrKeyNotFound {

					km.logger.Info().Msgf("secret ssh %s error %s", secret.Uuid, err)
					err := km.generateSecretBytes(secret, tx)
					if err != nil {
						km.logger.Info().Msgf("secret ssh %s entropy collection error %s", secret.Uuid, err)
					} else {
						km.logger.Info().Str("secret ssh %s entropy success", secret.Uuid)

					}
				}
			}
		} else if secret.IsEnv() {
			if _, err := ss.Get(secret.Uuid, secret.Key); err != nil {
				if err == stores.ErrKeyNotFound {

					km.logger.Info().Msgf("secret env %s rekey", secret.Uuid)
					err := ss.Set(secret.Uuid, secret.Key, []byte(`{"value": "MISSING"}`))
					if err != nil {
						km.logger.Info().Msgf("secret env %s rekey error %s", secret.Uuid, err)
					} else {
						km.logger.Info().Msgf("secret env %s rekey done", secret.Uuid)
					}
				}
			}
		}
	}

	tx.Commit()
	return nil
}

func Main() {

	var logger zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

	c := config.GetConfig()
	redisClient := redis.NewTCPClient(c.RedisConnOpts(1))

	km := keymaker{
		logger:      logger,
		config:      c,
		redisClient: redisClient,
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "rekey":
			err := km.rekey()
			if err != nil {
				logger.Fatal().Err(err)
			}
		case "rekey-missing":
			err := km.rekeyMissing()
			if err != nil {
				logger.Fatal().Err(err)
			}
		}
		os.Exit(0)
	}

	activityBus := activity.NewAMQPTransport(c.AmqpConnectionString(), "keymaker")
	defer activityBus.Close()
	bus := broadcast.NewAMQPTransport(c.AmqpConnectionString(), "keymaker")
	defer bus.Close()
	work, err := bus.Consume(broadcast.Create)
	if err != nil {
		logger.Fatal().Msgf("bus.consume(broadcast.create): %s", err)
	}

	for message := range work {
		switch message.Table() {
		case "repositories":
			go km.generateRepositoryCredential(activityBus, message)
		case "secrets":
			go km.generateSecret(c, redisClient, message)
		default:
			message.RejectForever()
		}
	}
}
