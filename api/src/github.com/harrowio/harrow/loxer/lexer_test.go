package loxer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func prettify(e Event) string {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func dumpEvent(e Event) {
	fmt.Fprintln(os.Stderr, prettify(e))
}

func recordEvents(result *[]Event) func(Event) {
	return func(e Event) {
		*result = append(*result, e)
	}
}

type testCase struct {
	input    string
	expected []Event
}

func marshalEqual(e1, e2 Event) bool {
	b1, _ := json.Marshal(e1)
	b2, _ := json.Marshal(e2)
	return bytes.Equal(b1, b2)
}

func (tc testCase) run(t *testing.T) {
	events := []Event{}
	l := newLexer(recordEvents(&events))
	l.feed([]byte(tc.input))
	// t.Logf("INPUT %q\n", tc.input)
	l.Close()

	if got, expected := len(events), len(tc.expected); got != expected {
		t.Errorf("Expected %d events, got %d", expected, got)
		if got > expected {
			t.Errorf("Extra events:\n")
			extra := events[expected:]
			for _, e := range extra {
				t.Errorf("%s\n", prettify(e))
			}
		}
	}

	for i, e := range events {
		if i >= len(tc.expected) {
			continue
		}
		if expected := tc.expected[i]; !marshalEqual(e, expected) {
			t.Errorf("\n%-10s\n%s\n%-10s\n%s\n",
				"Expected",
				prettify(expected),
				"Got",
				prettify(e),
			)
			// } else {
			// 	t.Logf("EVENT %s\n", reflect.ValueOf(e).Elem().FieldByName("Type"))
		}
	}
}

type testCases []testCase

func (tcs testCases) run(t *testing.T) {
	for _, tc := range tcs {
		tc.run(t)
	}
}

func Test_lexer_emit(t *testing.T) {
	spec := testCases{
		{
			"ab\tc",
			[]Event{
				&TextEvent{EventData{"text", "ab", 0}},
				&CursorEvent{
					EventData: EventData{"cursor", "\t", 2},
					Action:    CursorHorizontalTab,
				},
				&TextEvent{EventData{"text", "c", 3}},
			},
		}, {
			"abc",
			[]Event{
				&TextEvent{EventData{"text", "abc", 0}},
			},
		}, {
			"a\033[1;31mb\033[0m",
			[]Event{
				&TextEvent{EventData{"text", "a", 0}},
				&DisplayEvent{
					EventData:  EventData{"display", "\033[1;31m", 1},
					Attribute:  DisplayBright,
					Foreground: 2,
				},
				&TextEvent{EventData{"text", "b", 8}},
				&DisplayEvent{
					EventData: EventData{"display", "\033[0m", 9},
					Attribute: DisplayReset,
				},
			},
		}, {
			"\033[1A",
			[]Event{
				&CursorEvent{
					EventData: EventData{"cursor", "\033[1A", 0},
					Action:    CursorUp,
					Count:     1,
				},
			},
		}, {
			// Extended foreground colors
			"\033[38;5;127m",
			[]Event{
				&DisplayEvent{
					EventData:  EventData{"display", "\033[38;5;127m", 0},
					Foreground: 128,
					Attribute:  DisplayNone,
				},
			},
		}, {
			// Extended background colors
			"\033[48;5;127m",
			[]Event{
				&DisplayEvent{
					EventData:  EventData{"display", "\033[48;5;127m", 0},
					Background: 128,
					Attribute:  DisplayNone,
				},
			},
		}, {
			// Color codes as produced by capistrano
			"\033[0;30;49m",
			[]Event{
				&DisplayEvent{
					EventData:  EventData{"display", "\033[0;30;49m", 0},
					Background: 10,
					Foreground: 1,
					Attribute:  DisplayReset,
				},
			},
		}, {
			// Example output from capistrano
			"\x1B[0;30;49mDEBUG\x1B[0m [\x1B[0;32;49m0e04ec0e\x1B[0m] Command: \x1B[0;34;49m/usr/bin/env mkdir -p /tmp/railstutorialapp/\x1B[0m\n",
			[]Event{
				&DisplayEvent{
					EventData:  EventData{"display", "\033[0;30;49m", 0},
					Background: 10,
					Foreground: 1,
					Attribute:  DisplayReset,
				},
				&TextEvent{
					EventData: EventData{"text", "DEBUG", len("\033[0;30;49m")},
				},
				&DisplayEvent{
					EventData: EventData{"display", "\033[0m", len("\x1B[0;30;49mDEBUG")},
					Attribute: DisplayReset,
				},
				&TextEvent{
					EventData: EventData{"text", " [", len("\x1B[0;30;49mDEBUG\x1B[0m")},
				},
				&DisplayEvent{
					EventData:  EventData{"display", "\033[0;32;49m", len("\x1B[0;30;49mDEBUG\x1B[0m [")},
					Attribute:  DisplayReset,
					Foreground: 3,
					Background: 10,
				},
				&TextEvent{
					EventData: EventData{"text", "0e04ec0e", len("\x1B[0;30;49mDEBUG\x1B[0m [\033[0;32;49m")},
				},
				&DisplayEvent{
					EventData: EventData{"display", "\033[0m", len("\x1B[0;30;49mDEBUG\x1B[0m [\033[0;32;49m0e04ec0e")},
					Attribute: DisplayReset,
				},
				&TextEvent{
					EventData: EventData{"text", "] Command: ", len("\x1B[0;30;49mDEBUG\x1B[0m [\033[0;32;49m0e04ec0e\033[0m")},
				},
				&DisplayEvent{
					EventData:  EventData{"display", "\033[0;34;49m", len("\x1B[0;30;49mDEBUG\x1B[0m [\033[0;32;49m0e04ec0e\033[0m] Command: ")},
					Attribute:  DisplayReset,
					Foreground: 5,
					Background: 10,
				},
				&TextEvent{
					EventData: EventData{"text", "/usr/bin/env mkdir -p /tmp/railstutorialapp/", len("\x1B[0;30;49mDEBUG\x1B[0m [\033[0;32;49m0e04ec0e\033[0m] Command: \033[0;34;49m")},
				},
				&DisplayEvent{
					EventData: EventData{"display", "\033[0m", len("\x1B[0;30;49mDEBUG\x1B[0m [\033[0;32;49m0e04ec0e\033[0m] Command: \033[0;34;49m/usr/bin/env mkdir -p /tmp/railstutorialapp/")},
				},
				&CursorEvent{
					EventData: EventData{"cursor", "\n", len("\x1B[0;30;49mDEBUG\x1B[0m [\033[0;32;49m0e04ec0e\033[0m] Command: \033[0;34;49m/usr/bin/env mkdir -p /tmp/railstutorialapp/\033[0m")},
					Action:    CursorLineFeed,
				},
			},
		}, {
			"\033]10;\"fold-title\"\a",
			[]Event{
				&FoldEvent{
					EventData: EventData{"fold", "\033]10;\"fold-title\"\a", 0},
					Action:    FoldOpen,
					Title:     "fold-title",
				},
			},
		}, {
			"\033]10\a",
			[]Event{
				&FoldEvent{
					EventData: EventData{"fold", "\033]10\a", 0},
					Action:    FoldClose,
				},
			},
		},
	}

	spec.run(t)
}

func Test_parseEventArgs(t *testing.T) {
	testcases := []struct {
		input    string
		expected *eventArgs
	}{
		{"\033[10A", &eventArgs{numeric: []int{10}}},
		{"\033[10;12;14m", &eventArgs{numeric: []int{10, 12, 14}}},
		{"\033[10;\"hello-world\"m", &eventArgs{numeric: []int{10}, str: "hello-world"}},
		{"\033[0;34;49m", &eventArgs{numeric: []int{0, 34, 49}}},
		{"\033[10;\"hello \\\"world\\\"\"m", &eventArgs{numeric: []int{10}, str: `hello "world"`}},
	}

	for _, testcase := range testcases {
		result := parseEventArgs(testcase.input)
		if !reflect.DeepEqual(result, testcase.expected) {
			t.Errorf("Input: %q\nGot: %#v\nWant: %#v\n", testcase.input, result, testcase.expected)
		}
	}
}
