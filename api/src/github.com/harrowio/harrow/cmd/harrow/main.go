package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/harrowio/harrow/cmd/activity-worker"
	"github.com/harrowio/harrow/cmd/api"
	"github.com/harrowio/harrow/cmd/build-status-worker"
	controllerLXD "github.com/harrowio/harrow/cmd/controller-lxd"
	controllerShell "github.com/harrowio/harrow/cmd/controller-shell"
	"github.com/harrowio/harrow/cmd/fsbuilder"
	"github.com/harrowio/harrow/cmd/git-trigger-worker"
	"github.com/harrowio/harrow/cmd/harrow-archivist"
	"github.com/harrowio/harrow/cmd/harrow-update-repository-metadata"
	"github.com/harrowio/harrow/cmd/keymaker"
	limits "github.com/harrowio/harrow/cmd/limits"
	mail "github.com/harrowio/harrow/cmd/mail"
	"github.com/harrowio/harrow/cmd/mail-dispatcher"
	"github.com/harrowio/harrow/cmd/metadata-preflight"
	"github.com/harrowio/harrow/cmd/migrate"
	"github.com/harrowio/harrow/cmd/notifier"
	"github.com/harrowio/harrow/cmd/op-metrics"
	"github.com/harrowio/harrow/cmd/postal-worker"
	"github.com/harrowio/harrow/cmd/projector"
	"github.com/harrowio/harrow/cmd/report-build-status-to-github"
	"github.com/harrowio/harrow/cmd/runner"
	"github.com/harrowio/harrow/cmd/scheduler"
	"github.com/harrowio/harrow/cmd/user-script-runner"
	"github.com/harrowio/harrow/cmd/ws"
	"github.com/harrowio/harrow/cmd/zob"
	"github.com/rs/zerolog"
)

func main() {

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	programs := map[string]func(){
		activityWorker.ProgramName:                 activityWorker.Main,
		api.ProgramName:                            api.Main,
		buildStatusWorker.ProgramName:              buildStatusWorker.Main,
		opMetrics.ProgramName:                      opMetrics.Main,
		controllerLXD.ProgramName:                  controllerLXD.Main,
		controllerShell.ProgramName:                controllerShell.Main,
		fsbuilder.ProgramName:                      fsbuilder.Main,
		gitTriggerWorker.ProgramName:               gitTriggerWorker.Main,
		harrowArchivist.ProgramName:                harrowArchivist.Main,
		limits.ProgramName:                         limits.Main,
		mail.ProgramName:                           mail.Main,
		harrowUpdateRepositoryMetadata.ProgramName: harrowUpdateRepositoryMetadata.Main,
		keymaker.ProgramName:                       keymaker.Main,
		mailDispatcher.ProgramName:                 mailDispatcher.Main,
		metadataPreflight.ProgramName:              metadataPreflight.Main,
		migrate.ProgramName:                        migrate.Main,
		notifier.ProgramName:                       notifier.Main,
		runner.ProgramName:                         runner.Main,
		postalWorker.ProgramName:                   postalWorker.Main,
		projector.ProgramName:                      projector.Main,
		reportBuildStatusToGitHub.ProgramName:      reportBuildStatusToGitHub.Main,
		scheduler.ProgramName:                      scheduler.Main,
		userScriptRunner.ProgramName:               userScriptRunner.Main,
		ws.ProgramName:                             ws.Main,
		zob.ProgramName:                            zob.Main,
	}

	programName := filepath.Base(os.Args[0])

	program, found := programs[programName]
	if !found && len(os.Args) > 1 {
		programName = filepath.Base(os.Args[1])
		os.Args = os.Args[1:]
	}

	program, found = programs[programName]
	if !found {
		logger.Fatal().Msgf("unknown program: %s", programName)
	}

	program()
}

func programInfo(program func()) string {
	function := runtime.FuncForPC(reflect.ValueOf(program).Pointer())
	filename, line := function.FileLine(function.Entry())
	return fmt.Sprintf("%s:%d", filename, line)
}
