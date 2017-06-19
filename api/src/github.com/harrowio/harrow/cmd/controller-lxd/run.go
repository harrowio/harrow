package controllerLXD

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	redis "gopkg.in/redis.v2"

	"github.com/jmoiron/sqlx"

	"golang.org/x/crypto/ssh"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/loxer"

	"github.com/harrowio/harrow/bus/logevent"
	"github.com/harrowio/harrow/cast"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type FatalError struct {
	error
}

func runUserScript(log logger.Logger, client *ssh.Client, activitySink ActivitySink, db *sqlx.DB, operationUuid string, entrypoint string, config *config.Config) error {

	wg := new(sync.WaitGroup)
	logSinkClient := redis.NewTCPClient(config.RedisConnOpts(0))
	defer logSinkClient.Close()
	logSink := logevent.NewRedisTransport(logSinkClient, log)
	defer logSink.Close()

	controlMessages := make(chan *cast.ControlMessage, 0)
	lexemes := make(chan loxer.LexerEvent, 0)
	stdoutLoxer := loxer.NewLexer(func(e loxer.Event) {
		lexemes <- loxer.LexerEvent{
			T:     time.Now().UnixNano(),
			Fd:    1,
			Event: e,
		}
	})
	stderrLoxer := loxer.NewLexer(func(e loxer.Event) {
		lexemes <- loxer.LexerEvent{
			T:     time.Now().UnixNano(),
			Fd:    2,
			Event: e,
		}
	})
	loxerEvents := make([]*logevent.Message, 0)
	wg.Add(1)
	go func(log logger.Logger) {
		defer wg.Done()
		for l := range lexemes {
			err := logSink.Publish(operationUuid, int(l.Fd), l.T, l.Event)
			if err != nil {
				log.Warn().Msgf("unable to publish lexeme: %s", l)
			}
			loxerEvents = append(loxerEvents, &logevent.Message{
				O:  operationUuid,
				FD: int(l.Fd),
				T:  l.T,
				E:  loxer.SerializedEvent{Inner: l.Event},
			})
		}
		log.Debug().Msg("lexemes closed, persisting logs")
		ft := logevent.NewFileTransport(config, log)
		if err := ft.WriteLexemes(operationUuid, loxerEvents); err != nil {
			log.Error().Msgf("ft.writelexemes(): %s", err)
		} else {
			err := logSink.Expire(operationUuid)
			if err != nil {
				log.Error().Msgf("logsink.expire(%s): %s", operationUuid, err)
			}
			err = logSink.EOF(operationUuid)
			if err != nil {
				log.Error().Msgf("logsink.eof(%s): %s", operationUuid, err)
			}
		}
	}(log)
	wg.Add(1)
	go func(log logger.Logger) {
		defer wg.Done()
		for controlMessage := range controlMessages {
			switch controlMessage.Type {
			case cast.ChildExited:
				tx := mustBeginTx(db)
				defer tx.Rollback()
				close(lexemes)
				store := stores.NewDbOperationStore(tx)
				publish := &domain.Activity{}
				markExitStatus(log, store, db, controlMessage.ExitStatus, operationUuid, &publish)
				if err := tx.Commit(); err != nil {
					log.Error().Msgf("unable to commit: %s", err)
				}
				if publish != nil {
					activitySink.EmitActivity(publish)
				}
			case cast.Event:
				if controlMessage.Payload.Get("type") == "output" {
					parseOutput(log, controlMessage, stdoutLoxer, stderrLoxer)
					continue
				}
				err := handleEvent(db, controlMessage, activitySink, operationUuid)
				if err != nil {
					log.Error().Msgf("unable to handle event: %s", err)
				}
			}
		}
		log.Debug().Msg("controlmessages closed")
	}(log)

	session, err := client.NewSession()
	if err != nil {
		fatalError := FatalError{fmt.Errorf("Unable to connect to vm: %s", err)}
		return fatalError
	}
	defer session.Close()

	session.Stdout = cast.NewControlMessageParser(controlMessages)

	err = session.Run(entrypoint)
	stdoutLoxer.Close()
	stderrLoxer.Close()
	wg.Wait()
	return err
}

func parseOutput(log logger.Logger, msg *cast.ControlMessage, stdoutLoxer io.Writer, stderrLoxer io.Writer) {
	switch msg.Payload.Get("channel") {
	case "stderr":
		fmt.Fprintf(stderrLoxer, "%s", msg.Payload.Get("text"))
	case "stdout":
		fmt.Fprintf(stdoutLoxer, "%s", msg.Payload.Get("text"))
	default:
		log.Error().Msgf("unknown log channel: %s", msg.Payload.Get("channel"))
	}

}

func handleEvent(db *sqlx.DB, controlMessage *cast.ControlMessage, activitySink ActivitySink, operationUuid string) error {

	event := controlMessage.Payload
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	operationStore := stores.NewDbOperationStore(tx)
	operation, err := operationStore.FindByUuid(operationUuid)
	if err != nil {
		return err
	}

	if event.Get("event") == "fatal" {
		errStr := fmt.Sprintf("Fatal: script=%s line=%s func=%q cmd=%q status=%s", event.Get("script"), event.Get("line"), event.Get("func"), event.Get("cmd"), event.Get("status"))
		return markFatal(db, operationUuid, errStr)
	}

	if event.Get("event") == "analyze-repository" {
		return markRepositoryMetadata(db, activitySink, event)
	}

	operation.HandleEvent(event)
	if err := operationStore.MarkStatusLogs(operationUuid, operation.StatusLogs); err != nil {
		return err
	}
	if err := operationStore.MarkRepositoryCheckouts(operationUuid, operation.RepositoryCheckouts); err != nil {
		return err
	}
	if err := operationStore.MarkGitLogs(operationUuid, operation.GitLogs); err != nil {
		return err
	}

	return tx.Commit()
}

func markRepositoryInaccesible(db *sqlx.DB, operationUuid string) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	operation, err := stores.NewDbOperationStore(tx).FindByUuid(operationUuid)
	if err != nil {
		return err
	}

	if operation.RepositoryUuid == nil {
		return nil
	}
	repositories := stores.NewDbRepositoryStore(tx)
	if err := repositories.MarkAsAccessible(*operation.RepositoryUuid, false); err != nil {
		return err
	}

	return tx.Commit()
}

func markExitStatus(log logger.Logger, store *stores.DbOperationStore, db *sqlx.DB, s int, operationUuid string, publish **domain.Activity) error {
	err := store.MarkExitStatus(operationUuid, s)
	if err != nil {
		return err
	}
	operation, err := store.FindByUuid(operationUuid)
	if err != nil {
		return err
	}

	if s != 0 {
		if operation.FatalError != nil {
			*publish = activities.OperationFailedFatally(operation)
		} else {
			*publish = activities.OperationFailed(operation)
		}
		if err := markRepositoryInaccesible(db, operationUuid); err != nil {
			log.Error().Msgf("markrepositoryinaccesible(db, operationuuid): %s\n", err)
		}

		return store.MarkAsFailed(operationUuid)
	} else {
		*publish = activities.OperationSucceeded(operation)
		return store.MarkAsFinished(operationUuid)
	}
}

func markRepositoryMetadata(db *sqlx.DB, activitySink ActivitySink, event domain.EventPayload) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	repositoryUuid := event.Get("repository")
	metadata := &domain.RepositoryMetaData{}
	if err := json.Unmarshal([]byte(event.Get("data")), metadata); err != nil {
		return err
	}

	store := stores.NewDbRepositoryStore(tx)
	repository, err := store.FindByUuid(repositoryUuid)
	if err != nil {
		return err
	}

	action := NewUpdateRepositoryMetaData(activitySink, store)
	if err := action.Update(repositoryUuid, repository.Metadata, metadata); err != nil {
		return err
	}
	return tx.Commit()
}
