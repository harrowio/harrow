// Package cast implements running other processes on a remote machine
// and harvesting their output via TCP.  The following mapping is in
// place for the subprocess:
//
//    stdin    /dev/null
//    stdout   localhost:2001
//    stderr   localhost:2002
//
// Additionally, control messages are sent on port 2003 to communicate
// the exit status of the subprocess.  A heartbeat message is sent
// periodically on this channel to determine whether the process
// should continue running.  If the heartbeat cannot be sent, the
// process is terminated with SIGKILL.
package cast

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"
)

// DefaultHeartrate is the interval between sending heartbeat messages
// if no interval is specified explicitly.
const (
	DefaultHeartrate = 1 * time.Minute
	ExitStatusNone   = -1
)

var (
	// ErrConnectionLost is returned if the connection on the
	// control channel has been lost.
	ErrConnectionLost = errors.New("cast: connection lost")

	// ErrTimedOut is returned if no connection could be
	// established after retrying.
	ErrTimedOut = errors.New("cast: timed out")
)

// Command represents an instrumented subprocess that should send its
// output over a socket.  The subprocess is killed if the connection
// to the controlling process is lost.
type Command struct {
	// cmd represents the subprocess to run.
	cmd *exec.Cmd

	// ctrlChannel is the connection over which to send control
	// messages.
	ctrlChannel net.Conn

	// basePort is used for calculating the port numbers.  The
	// ports used are basePort + fd number (+ 1 for the control
	// channel).
	basePort int16

	// closeAfterRun is a list of things that need to be closed
	// after the subprocess has finished running (for whatever
	// reason).
	closeAfterRun []io.Closer

	// heartrate is the interval between sending heartbeat
	// messages.
	heartrate time.Duration
}

// Run starts the subprocess and waits for it to finish.  If the
// connection on the control channel is lost, the subprocess is
// terminated.  Otherwise its exit status is reported on the control
// channel.
func (self *Command) Run() error {
	if err := self.retry(self.connect); err != nil {
		return err
	}

	defer self.close()

	done := make(chan error, 1)
	heartbeat := make(chan error, 1)
	go self.sendHeartbeats(heartbeat)
	go func() {
		if err := self.cmd.Run(); err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				done <- err
			}
		}

		done <- nil
	}()

	select {
	case <-done:
		if err := self.reportExitStatus(); err != nil {
			return err
		}
	case <-heartbeat:
		if err := self.cmd.Process.Kill(); err != nil {
			return err
		}

		return ErrConnectionLost
	}

	return nil
}

// retry tries to run a function until it does not return an error.
// It uses an exponentially increasing delay between tries and tries
// up to 3 times.
func (self *Command) retry(do func() error) error {
	maxRetries := 3
	retries := 0
	err := (error)(nil)
	delay := 2 * time.Second

	for retries < maxRetries {
		err = do()
		if err == nil {
			return nil
		}

		time.Sleep(delay)

		delay = delay * 2

		retries++
	}

	return ErrTimedOut
}

func (self *Command) sendHeartbeats(done chan error) {
	heartrate := time.Tick(self.heartrate)
	for range heartrate {
		err := self.Heartbeat()
		if err != nil {
			done <- ErrConnectionLost
		}
	}
}

func (self *Command) close() {
	for _, closer := range self.closeAfterRun {
		closer.Close()
	}
}

func (self *Command) reportExitStatus() error {
	return self.sendControlMessage(&ControlMessage{
		Type:       ChildExited,
		ExitStatus: self.ExitStatus(),
	})
}

// sendControlMessage serializes a control message as JSON and sends
// it over the control channel.
func (self *Command) sendControlMessage(msg *ControlMessage) error {
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}

	_, err = self.ctrlChannel.Write(append(data, '\n'))

	if err != nil {
		return err
	}

	return nil
}

// ExitStatus returns the exit status of the subprocess.  If the
// subprocess has not been started yet or is still running it returns
// ExitStatusNone.
func (self *Command) ExitStatus() int {
	return ExitStatusFor(self.cmd)
}

// Heartbeat sends a heartbeat message of the control channel.
func (self *Command) Heartbeat() error {
	return self.sendControlMessage(&ControlMessage{Type: Heartbeat})
}

// connect establishes the necessary socket connections for forwarding
// stdout and stderr.
func (self *Command) connect() error {
	if err := self.openControlChannel(); err != nil {
		return err
	}

	if err := self.connectStdout(); err != nil {
		return err
	}

	if err := self.connectStderr(); err != nil {
		return err
	}

	return nil
}

func (self *Command) openControlChannel() error {
	conn, err := self.connForFd(3)
	if err != nil {
		return err
	}

	self.ctrlChannel = conn
	return nil
}

func (self *Command) connectStdout() error {
	conn, err := self.connForFd(1)
	if err != nil {
		return err
	}

	self.cmd.Stdout = conn
	return nil
}

func (self *Command) connectStderr() error {
	conn, err := self.connForFd(2)
	if err != nil {
		return err
	}
	self.cmd.Stderr = conn
	return nil
}

func (self *Command) connForFd(fd int16) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", self.basePort+fd))
	if err != nil {
		return nil, err
	}

	self.closeAfterRun = append(self.closeAfterRun, conn)

	return conn, nil
}

// SetBasePort sets the base port to use for port number calculations.
// The following ports are in use:
//
//     stdout = base port + 1
//     stderr = base port + 2
//     ctrl   = base port + 3
//
// Ctrl is the channel used for sending control messages to the
// orchestrator.
//
// Calling this method after Run has been called has no effect.
func (self *Command) SetBasePort(port int) {
	self.basePort = int16(port)
}

// New creates a command for running prog with args.  It uses
// DefaultHeartrate as the interval between heartbeat messages.
func New(prog string, args ...string) *Command {
	return NewWithHeartrate(DefaultHeartrate, prog, args...)
}

// NewWithHeartrate creates a command for running prog with args,
// using heartrate as the interval between heartbeat messages.
func NewWithHeartrate(heartrate time.Duration, prog string, args ...string) *Command {
	cmd := exec.Command(prog, args...)
	cmd.Stdin = nil
	return &Command{
		cmd:       cmd,
		basePort:  2000,
		heartrate: heartrate,
	}
}
