package cast

import (
	"os/exec"
	"syscall"
)

func ExitStatusFor(cmd *exec.Cmd) int {
	state := cmd.ProcessState

	if state == nil || !state.Exited() {
		return ExitStatusNone
	}

	// this will only ever run on *NIX
	waitStatus := state.Sys().(syscall.WaitStatus)
	return waitStatus.ExitStatus()
}
