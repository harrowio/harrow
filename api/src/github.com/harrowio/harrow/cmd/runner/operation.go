package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type Operation struct {
	db       *sqlx.DB
	config   *config.Config
	lxd      *LXD
	log      logger.Logger
	reporter Reporter

	op *domain.Operation
}

func (o *Operation) RunLocally(notifierType string) {

	executable, _ := os.Executable()

	o.log.Info().Msg("running operation in local shell (lxd container wasn't necessary)")

	fsBuilder := exec.Command(
		executable,
		"fsbuilder",
		"-operation-uuid",
		o.op.Uuid,
	)

	controllerShell := exec.Command(
		executable,
		"controller-shell",
		"-operation-uuid",
		o.op.Uuid,
		fmt.Sprintf("/srv/harrow/bin/harrow-notify-%s", notifierType),
	)

	pr, pw := io.Pipe()
	fsBuilder.Stdout = pw
	fsBuilder.Stderr = os.Stdout

	controllerShell.Stdin = pr
	controllerShell.Stdout = os.Stdout
	controllerShell.Stderr = os.Stdout

	o.log.Info().Msgf(
		"about to run '%s %s' and pipe stdout to '%s %s'",
		fsBuilder.Path,
		strings.Join(fsBuilder.Args[1:], " "),
		controllerShell.Path,
		strings.Join(controllerShell.Args[1:], " "),
	)

	o.log.Debug().Msg("starting fsbuilder")
	if err := fsBuilder.Start(); err != nil {
		o.log.Error().Msgf("can't start fsbuilder %s", err)
		return
	}
	o.log.Debug().Msg("starting fsbuilder (done)")

	o.log.Debug().Msg("controller-lxd")
	if err := controllerShell.Start(); err != nil {
		o.log.Error().Msgf("can't start controller-lxd %s", err)
		return
	}
	o.log.Debug().Msg("controller-lxd (done)")

	go func(log logger.Logger) {
		defer pw.Close()
		if err := fsBuilder.Wait(); err != nil {
			log.Error().Msgf("fsbuilder exited with error: %s", err)
			return
		}
		log.Info().Msg("fsbuilder finished")
		pw.Close()
		log.Info().Msg("fsbuilder pipe closed")
	}(o.log)

	if err := controllerShell.Wait(); err != nil {
		o.log.Error().Msgf("controller-shell exited with error: %s", err)
		return
	}
	log.Info().Msg("controller shell finished")

	o.log.Info().Msg("operation finished, returning")

	return
}

func (o *Operation) RunOnLXDHost() {

	executable, _ := os.Executable()

	o.log.Info().Msg("running operation on lxd host")

	fsBuilder := exec.Command(
		executable,
		"fsbuilder",
		"-operation-uuid",
		o.op.Uuid,
	)

	controllerLXD := exec.Command(
		executable,
		"controller-lxd",
		"-operation-uuid",
		o.op.Uuid,
		"-entrypoint",
		fmt.Sprintf(`lxc exec %s -- sudo -u ubuntu -i user-script-runner-local .bin/setup`, o.lxd.containerName()),
		"-container-id",
		o.lxd.containerName(),
		"-connect",
		o.lxd.connURL.String(),
		"-operation-uuid",
		o.op.Uuid,
	)

	pr, pw := io.Pipe()
	fsBuilder.Stdout = pw
	fsBuilder.Stderr = os.Stdout

	controllerLXD.Stdin = pr
	controllerLXD.Stdout = os.Stdout
	controllerLXD.Stderr = os.Stdout

	o.log.Info().Msgf(
		"about to run '%s %s' and pipe stdout to '%s %s'",
		fsBuilder.Path,
		strings.Join(fsBuilder.Args[1:], " "),
		controllerLXD.Path,
		strings.Join(controllerLXD.Args[1:], " "),
	)

	o.log.Debug().Msg("starting fsbuilder")
	if err := fsBuilder.Start(); err != nil {
		o.log.Error().Msgf("can't start fsbuilder %s", err)
		return
	}
	o.log.Debug().Msg("starting fsbuilder (done)")

	o.log.Debug().Msg("controller-lxd")
	if err := controllerLXD.Start(); err != nil {
		o.log.Error().Msgf("can't start controller-lxd %s", err)
		return
	}
	o.log.Debug().Msg("controller-lxd (done)")

	go func(log logger.Logger) {
		defer pw.Close()
		if err := fsBuilder.Wait(); err != nil {
			log.Error().Msgf("fsbuilder exited with error: %s", err)
			return
		}
		log.Info().Msg("fsbuilder finished")
		pw.Close()
		log.Info().Msg("fsbuilder pipe closed")
	}(o.log)

	if err := controllerLXD.Wait(); err != nil {
		o.log.Error().Msgf("controller-lxd exited with error: %s", err)
		return
	}
	log.Info().Msg("controller lxd finished")

	o.log.Info().Msg("operation finished, returning")

	return
}
