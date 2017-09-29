package controllerShell

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/stores"
)

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

const ProgramName = "controller-shell"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	operationUuid := flag.String("operation-uuid", "", "The operation to run")

	flag.Parse()

	c := config.GetConfig()
	db, err := c.DB()
	if err != nil {
		log.Fatal().Msgf("unable to open db: %s", err)
	}

	defer db.Close()

	if *operationUuid == "" {
		fmt.Fprint(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "No command specified!\n")
		os.Exit(3)
	}

	if err := run(db, *operationUuid, flag.Args()); err != nil {
		log.Fatal().Msgf("run(): %s", err)
	}
}

func run(db *sqlx.DB, operationUuid string, args []string) error {

	commandToRun := exec.Command(args[0], args[1:]...)
	commandToRun.Stdin = os.Stdin
	out, err := commandToRun.CombinedOutput()
	log.Debug().Msgf("\nOUTPUT:\n%s\n", out)

	if err == nil {
		mustMarkExit(db, operationUuid, 0)
	} else if e, ok := err.(*exec.ExitError); ok {
		log.Info().Msgf("%s %#v", e, e)
		if status, ok := e.Sys().(syscall.WaitStatus); ok {
			log.Info().Msgf("%s %#v", status, status)
			mustMarkExit(db, operationUuid, status.ExitStatus())
		}
	} else {
		msg := fmt.Sprintf("%s: %s\n---\n%s\n", strings.Join(args, " "), err, out)
		mustMarkFatal(db, operationUuid, msg)
		return errors.New(msg)
	}

	return nil
}

func mustMarkStarted(db *sqlx.DB, operationUuid string) {
	tx := mustBeginTx(db)
	defer tx.Rollback()
	store := stores.NewDbOperationStore(tx)
	if err := store.MarkAsStarted(operationUuid); err != nil {
		log.Fatal().Msgf("Unable to mark started: %s", err)
	}
	mustCommitTx(tx)
}

func mustMarkExit(db *sqlx.DB, operationUuid string, status int) {
	tx := mustBeginTx(db)
	defer tx.Rollback()
	opStore := stores.NewDbOperationStore(tx)
	_, err := opStore.FindByUuid(operationUuid)
	if err != nil {
		log.Fatal().Msgf("Unable to load operation: %s", err)
	}

	if err := opStore.MarkExitStatus(operationUuid, status); err != nil {
		log.Fatal().Msgf("Unable to update exit status: %s", err)
	}
	if status == 0 {
		if err := opStore.MarkAsFinished(operationUuid); err != nil {
			log.Fatal().Msgf("Unable to mark finished: %s", err)
		}
	} else {
		if err := opStore.MarkAsFailed(operationUuid); err != nil {
			log.Fatal().Msgf("Unable to mark failed: %s", err)
		}
	}

	mustCommitTx(tx)
}

func mustMarkFatal(db *sqlx.DB, operationUuid string, fatal string) {
	tx := mustBeginTx(db)
	defer tx.Rollback()
	store := stores.NewDbOperationStore(tx)
	if err := store.MarkFatalError(operationUuid, fatal); err != nil {
		log.Fatal().Msgf("Unable to update fatal error: %s", err)
	}
	mustCommitTx(tx)
}

func mustBeginTx(db *sqlx.DB) *sqlx.Tx {
	tx, err := db.Beginx()
	if err != nil {
		log.Fatal().Msgf("Unable to begin tx: %s", err)
	}
	return tx
}

func mustCommitTx(tx *sqlx.Tx) {
	if err := tx.Commit(); err != nil {
		log.Fatal().Msgf("Unable to commit tx: %s", err)
	}
}
