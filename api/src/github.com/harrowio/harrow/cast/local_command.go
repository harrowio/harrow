package cast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// LocalCommand represents an instrumented command that writes a
// detailed, structured log of activities to stdout instead of sending
// it over multiple network sockets.

// The output is a JSON stream of ControlMessage objects.  Regular
// output of the process on stdout or stderr is captured in events of
// type "output", setting the "channel" field to "stdout" or "stderr"
// accordingly.
type LocalCommand struct {
	cmd *exec.Cmd
	out io.Writer
}

// NewLocalCommand constructs a new LocalCommand for running prog with args, sending
// control messages to out.
func NewLocalCommand(out io.Writer, prog string, args ...string) *LocalCommand {
	cmd := exec.Command(prog, args...)
	self := &LocalCommand{
		cmd: cmd,
		out: out,
	}
	self.connectStreamsToMessageEmitters(cmd, out)
	return self
}

// Connects the command's stdout and stderr to a write which converts
// any writes into ControlMessage objects.
func (self *LocalCommand) connectStreamsToMessageEmitters(cmd *exec.Cmd, out io.Writer) {
	cmd.Stdout = NewControlMessageWriter("stdout", out)
	cmd.Stderr = NewControlMessageWriter("stderr", out)
}

// String returns the string representation of the underlying shell
// command in a syntax suitable for use with a shell.
//
//     c := NewLocalCommand(os.Stdout, "bash", "-c", `printf "hello\n"`)
//     c.String() // bash -c "printf \"hello\\n""
func (self *LocalCommand) String() string {
	out := bytes.NewBufferString(self.cmd.Args[0])
	for _, arg := range self.cmd.Args[1:] {
		if strings.ContainsAny(arg, " \n\t\r'\"") {
			fmt.Fprintf(out, " %q", arg)
		} else {
			fmt.Fprintf(out, " %s", arg)
		}
	}

	return out.String()
}

// Run runs the underlying shell command, waiting for it to complete.
func (self *LocalCommand) Run() error {
	readControlMessages, writeControlMessages, err := os.Pipe()
	if err != nil {
		return err
	}

	go io.Copy(self.out, readControlMessages)

	self.cmd.ExtraFiles = []*os.File{writeControlMessages}
	err = self.cmd.Run()
	time.Sleep(10 * time.Millisecond)
	self.reportExitStatus()
	return err
}

// reportExitStatus writes a control message reporting the exit status
// contained in exitErr.  If exitErr is nil, an exit status of 0 is
// reported.
func (self *LocalCommand) reportExitStatus() {
	exitStatus := ExitStatusFor(self.cmd)
	json.NewEncoder(self.out).Encode(&ControlMessage{
		Type:       ChildExited,
		ExitStatus: exitStatus,
		Payload:    Payload{},
	})
}

// ExitStatus returns the exit status of running the command.  It
// returns -1 if ExitStatus is called before calling Run.
func (self *LocalCommand) ExitStatus() int {
	if self.cmd.ProcessState.Exited() {
		return ExitStatusFor(self.cmd)
	}

	return -1
}
