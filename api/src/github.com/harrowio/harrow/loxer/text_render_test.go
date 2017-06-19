package loxer

import "testing"

func TestTextRenderer_renders_text_events(t *testing.T) {
	subject := NewTextRenderer()
	event := NewTextEvent("the text", 0)
	subject.Handle(event)
	if got, want := subject.String(), "the text"; got != want {
		t.Fatalf("subject.String() = %#v; want %#v", got, want)
	}
}

func TestTextRenderer_renders_text_multiple_events(t *testing.T) {
	subject := NewTextRenderer()
	event := NewTextEvent("the text", 0)
	subject.Handle(event)
	subject.Handle(event)
	if got, want := subject.String(), "the textthe text"; got != want {
		t.Fatalf("subject.String() = %#v; want %#v", got, want)
	}
}

func TestTextRenderer_renders_cursor_events(t *testing.T) {
	subject := NewTextRenderer()
	event := NewCursorEvent("\n", 0)
	subject.Handle(event)
	if got, want := subject.String(), "\n"; got != want {
		t.Fatalf("subject.String() = %#v; want %#v", got, want)
	}
}

func TestTextRenderer_renders_a_tab_as_two_spaces_or_three_to_piss_paul_off(t *testing.T) {
	subject := NewTextRenderer()
	event := NewCursorEvent("\t", 0)
	subject.Handle(event)
	if got, want := subject.String(), "  "; got != want {
		t.Fatalf("subject.String() = %#v; want %#v", got, want)
	}
}

func TestTextRenderer_renders_many_events(t *testing.T) {
	subject := NewTextRenderer()
	textevent := NewTextEvent("the text", 0)
	cursorevent := NewCursorEvent("\n", 0)
	subject.Handle(textevent)
	subject.Handle(cursorevent)
	subject.Handle(textevent)
	if got, want := subject.String(), "the text\nthe text"; got != want {
		t.Fatalf("subject.String() = %#v; want %#v", got, want)
	}
}
