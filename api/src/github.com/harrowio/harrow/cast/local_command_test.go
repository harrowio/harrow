package cast

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"
)

func TestLocalCommand_Run_runs_command_writing_control_messages_to_stdout(t *testing.T) {
	out := bytes.NewBufferString("")
	cmd := NewLocalCommand(out, "bash", "-c", `printf "stdout\n"; sleep 1; printf "stderr\n" >&2`)
	if err := cmd.Run(); err != nil {
		t.Fatalf("run %s: %s", cmd, err)
	}

	stdout := &ControlMessage{}
	stderr := &ControlMessage{}
	dec := json.NewDecoder(out)

	if err := dec.Decode(stdout); err != nil {
		t.Fatal("stdout", err)
	}

	if err := dec.Decode(stderr); err != nil {
		t.Fatal("stderr", err)
	}

	if got, want := stdout.Payload.Get("text"), "stdout\n"; got != want {
		t.Errorf(`stdout.Payload.Get("text") = %q; want %q`, got, want)
	}

	if got, want := stderr.Payload.Get("text"), "stderr\n"; got != want {
		t.Errorf(`stderr.Payload.Get("text") = %q; want %q`, got, want)
	}
}

func TestLocalCommand_Run_emits_child_exited_message_after_command_has_finished(t *testing.T) {
	out := bytes.NewBufferString("")
	cmd := NewLocalCommand(out, "bash", "-c", `exit 1`)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("run %s: %s", cmd, err)
		}
	}

	exited := &ControlMessage{}
	dec := json.NewDecoder(out)
	if err := dec.Decode(exited); err != nil {
		t.Fatal("exited", err)
	}

	if got, want := exited.Type, ChildExited; got != want {
		t.Errorf(`exited.Type = %v; want %v`, got, want)
	}

	if got, want := exited.ExitStatus, 1; got != want {
		t.Errorf(`exited.ExitStatus = %v; want %v`, got, want)
	}
}

func TestLocalCommand_Run_accepts_control_messages_on_fd_3_in_the_command(t *testing.T) {
	out := bytes.NewBufferString("")
	cmd := NewLocalCommand(out, "bash", "-c", `printf '{"type":"test"}' >&3`)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("run %s: %s", cmd, err)
		}
	}

	test := &ControlMessage{}
	dec := json.NewDecoder(out)
	if err := dec.Decode(test); err != nil {
		t.Fatal("test", err)
	}

	if got, want := test.Type, MessageType("test"); got != want {
		t.Errorf(`test.Type = %v; want %v`, got, want)
	}
}

func TestLocalCommand_ExitStatus_reports_the_exit_status_after_calling_Run(t *testing.T) {
	out := bytes.NewBufferString("")
	cmd := NewLocalCommand(out, "bash", "-c", `exit 123`)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("run %s: %s", cmd, err)
		}
	}

	if got, want := cmd.ExitStatus(), 123; got != want {
		t.Errorf(`cmd.ExitStatus() = %v; want %v`, got, want)
	}
}
