package operationRunner

import (
	"fmt"
	"os"
	"syscall"

	"github.com/harrowio/harrow/config"
)

const (
	userJobTmplLXD = `%s fsbuilder -operation-uuid %s | %s controller-lxd -entrypoint "lxc exec %s -- sudo -u ubuntu -i user-script-runner-local .bin/setup" -container-id %s -connect %s -operation-uuid %s; %s vmex-lxd -container-id %s -connect %s`

	notifierJobTemplate = "%s fsbuilder -operation-uuid %s | %s controller-shell -operation-uuid %s '/srv/harrow/bin/harrow-notify-${NOTIFIER_TYPE}'"
)

func spawnUserJobLXD(c *config.Config, uuid string, id, ip string) error {

	cmdline := fmt.Sprintf(userJobTmplLXD, os.Args[0], uuid, os.Args[0], id, id, ip, uuid, os.Args[0], id, ip)
	argv := []string{"/bin/sh", "-c", cmdline}
	if err := startProcess(argv); err != nil {
		return fmt.Errorf("spawnUserJobLXD: %s\nCommand: %s\n", err, argv)
	}
	return nil
}

func localSpawn(command, operationUuid string) error {
	cmdline := fmt.Sprintf(command, os.Args[0], operationUuid, os.Args[0], operationUuid)
	argv := []string{"/bin/sh", "-c", cmdline}
	err := startProcess(argv)
	if err != nil {
		return fmt.Errorf("startProcess: %s", err)
	}
	return nil
}

func spawnNotifierJob(uuid string, notifierType string) error {

	command := os.Expand(notifierJobTemplate, func(key string) string {
		switch key {
		case "NOTIFIER_TYPE":
			return notifierType
		default:
			panic(fmt.Sprintf("unknown variable %q in %q", key, notifierJobTemplate))
		}
	})
	return localSpawn(command, uuid)
}

func startProcess(argv []string) error {

	var attr = os.ProcAttr{
		Dir: ".",
		Env: os.Environ(),
		Files: []*os.File{
			nil,
			os.Stdout,
			os.Stderr,
		},
		Sys: &syscall.SysProcAttr{Setsid: true},
	}

	process, err := os.StartProcess(argv[0], argv, &attr)
	if err != nil {
		return err
	}
	_, err = process.Wait()
	return err
}
