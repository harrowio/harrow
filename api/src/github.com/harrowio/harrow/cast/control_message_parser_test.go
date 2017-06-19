package cast

import (
	"fmt"
	"testing"
	"time"
)

func TestControlMessageParser_converts_bytes_to_control_messages(t *testing.T) {
	messages := make(chan *ControlMessage, 1)
	roundTrip := NewControlMessageWriter("stdout", NewControlMessageParser(messages))

	fmt.Fprintf(roundTrip, "a")

	message := <-messages
	if got, want := message.Type, Event; got != want {
		t.Errorf(`message.Type = %v; want %v`, got, want)
	}

	if got, want := message.Payload.Get("text"), "a"; got != want {
		t.Errorf(`message.Payload.Get("text") = %v; want %v`, got, want)
	}
}

func TestControlMessageParser_parses_messages_that_are_split_over_two_writes(t *testing.T) {
	messages := make(chan *ControlMessage, 2)
	parser := NewControlMessageParser(messages)
	fmt.Fprintf(parser, `{"type":`)
	fmt.Fprintf(parser, `"child_exited"}`)

	message := <-messages
	if got, want := message.Type, ChildExited; got != want {
		t.Errorf(`message.Type = %v; want %v`, got, want)
	}
}

func TestControlMessageParser_closes_channel_after_having_parsed_a_child_exited_message(t *testing.T) {
	messages := make(chan *ControlMessage, 2)
	parser := NewControlMessageParser(messages)
	fmt.Fprintf(parser, `{"type":`)
	fmt.Fprintf(parser, `"child_exited"}`)

	message := <-messages
	if got, want := message.Type, ChildExited; got != want {
		t.Errorf(`message.Type = %v; want %v`, got, want)
	}

	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out")
	case _, open := <-messages:
		if got, want := open, false; got != want {
			t.Errorf(`open = %v; want %v`, got, want)
		}
	}
}

func TestControlMessageParser_handles_two_messages_in_a_single_write(t *testing.T) {
	messages := make(chan *ControlMessage, 2)
	parser := NewControlMessageParser(messages)
	fmt.Fprintf(parser, `{"type":"a"}{"type":"b"}`)

	select {
	case a := <-messages:
		if got, want := a.Type, MessageType("a"); got != want {
			t.Errorf(`a.Type = %v; want %v`, got, want)
		}
		select {
		case b := <-messages:
			if got, want := b.Type, MessageType("b"); got != want {
				t.Errorf(`b.Type = %v; want %v`, got, want)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatalf("timed out for b")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("timed out for a")
	}
}

func TestControlMessageParser_emits_messages_as_soon_as_they_are_parsed(t *testing.T) {
	messages := make(chan *ControlMessage, 2)
	parser := NewControlMessageParser(messages)
	fmt.Fprintf(parser, `{"type":"a"}`)
	message := <-messages
	if got, want := message.Type, MessageType("a"); got != want {
		t.Errorf(`message.Type = %v; want %v`, got, want)
	}
	fmt.Fprintf(parser, `{"type":"b"}`)
	message = <-messages
	if got, want := message.Type, MessageType("b"); got != want {
		t.Errorf(`message.Type = %v; want %v`, got, want)
	}
}
