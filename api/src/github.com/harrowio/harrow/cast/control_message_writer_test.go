package cast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestControlMessageWriter_writes_control_message_to_out_for_every_write(t *testing.T) {
	out := bytes.NewBufferString("")
	w := NewControlMessageWriter("stdout", out)

	fmt.Fprintf(w, "Hello, world!\n")
	fmt.Fprintf(w, "Another write\n")

	dec := json.NewDecoder(out)

	helloWorld := &ControlMessage{}
	if err := dec.Decode(helloWorld); err != nil {
		t.Fatalf("decode helloWorld: %s", err)
	}

	anotherWrite := &ControlMessage{}
	if err := dec.Decode(anotherWrite); err != nil {
		t.Fatalf("decode anotherWrite: %s", err)
	}

	verify := func(msg *ControlMessage, tag, text string) {
		if got, want := msg.Type, Event; got != want {
			t.Errorf(`msg.Type = %v; want %v`, got, want)
		}

		if got, want := msg.Payload.Get("type"), "output"; got != want {
			t.Errorf(`msg.Payload.Get("type") = %v; want %v`, got, want)
		}

		if got, want := msg.Payload.Get("channel"), tag; got != want {
			t.Errorf(`msg.Payload.Get("channel") = %v; want %v`, got, want)
		}

		if got, want := msg.Payload.Get("text"), text; got != want {
			t.Errorf(`msg.Payload.Get("text") = %v; want %v`, got, want)
		}
	}

	verify(helloWorld, "stdout", "Hello, world!\n")
	verify(anotherWrite, "stdout", "Another write\n")
}
