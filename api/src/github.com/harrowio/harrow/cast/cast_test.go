package cast

import (
	"encoding/json"
	"log"
	"testing"
	"time"
)

var (
	stdoutSink  = NewSink(":22001")
	stderrSink  = NewSink(":22002")
	sysctrlSink = NewSink(":22003")
	sinks       = []*Sink{stdoutSink, stderrSink, sysctrlSink}
)

func init() {
	for _, sink := range sinks {
		if err := sink.Start(); err != nil {
			log.Fatal(err)
		}
	}
}

func withFailingHeartbeat(do func()) {
	// the control connection server closes the connection
	// immediately, so the heartbeat will be sent when the
	// connection has been closed already.
	savedTimeout := sysctrlSink.timeout
	sysctrlSink.timeout = 0
	defer func() { sysctrlSink.timeout = savedTimeout }()
	do()
}

func resetAll() {
	for _, sink := range sinks {
		sink.Reset()
	}
}

func Test_Cast_forwardsStdoutToSocket(t *testing.T) {
	defer resetAll()
	cmd := New("echo", "test-stdout")
	cmd.SetBasePort(22000)
	if err := cmd.Run(); err != nil {
		t.Errorf("cmd.Run: %s", err)
	}

	buf := stdoutSink.Recv()
	if got, want := buf.String(), "test-stdout\n"; got != want {
		t.Errorf("stdoutSink.received.String() = %q; want %q", got, want)
	}
}

func Test_Cast_reportsTheExitStatusOnTheSystemCtrlChannel(t *testing.T) {
	defer resetAll()

	cmd := New("sh", "-c", "exit 13")
	cmd.SetBasePort(22000)
	if err := cmd.Run(); err != nil {
		t.Errorf("cmd.Run: %s", err)
	}

	buf := sysctrlSink.Recv()

	ctrlMessage := &ControlMessage{}
	if err := json.Unmarshal(buf.Bytes(), ctrlMessage); err != nil {
		t.Fatal(err)
	}

	if got, want := ctrlMessage.ExitStatus, 13; got != want {
		t.Errorf("ctrlMessage.ExitStatus = %d; want %d", got, want)
	}
}

func Test_Cast_forwardsStderrToSocket(t *testing.T) {
	t.Skip("fails randomly")
	defer resetAll()
	cmd := New("sh", "-c", "echo test-stderr 1>&2")
	cmd.SetBasePort(22000)
	if err := cmd.Run(); err != nil {
		t.Errorf("cmd.Run: %s", err)
	}

	buf := stderrSink.Recv()
	if got, want := buf.String(), "test-stderr\n"; got != want {
		t.Errorf("stderrSink.received.String() = %q; want %q", got, want)
	}
}

func Test_Cast_sendsHeartbeatsOnTheSystemCtrlChannel(t *testing.T) {
	defer resetAll()
	cmd := NewWithHeartrate(10*time.Millisecond, "sh", "-c", "sleep 1; date")
	cmd.SetBasePort(22000)
	cmd.Run()

	buf := sysctrlSink.Recv()
	ctrlMessage := &ControlMessage{}
	if err := json.NewDecoder(buf).Decode(ctrlMessage); err != nil {
		t.Fatal(err)
	}

	if got, want := ctrlMessage.Type, Heartbeat; got != want {
		t.Errorf("ctrlMessage.Type = %q; want %q", got, want)
	}
}

func Test_Cast_killsChildProcess_ifHeartbeatFails(t *testing.T) {
	defer resetAll()

	do := func() {
		cmd := NewWithHeartrate(10*time.Millisecond, "sh", "-c", `trap "echo killed" KILL; sleep 2; echo survived`)
		cmd.SetBasePort(22000)
		if got, want := cmd.Run(), ErrConnectionLost; got != want {
			t.Errorf("cmd.Run: err = %s; want %s", got, want)
		}
	}

	withFailingHeartbeat(do)
}
