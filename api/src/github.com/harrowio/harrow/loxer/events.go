package loxer

import (
	"encoding/json"
	"fmt"
	"strings"
)

type LexerEvent struct {
	Fd uintptr
	T  int64
	Event
}

func (self *LexerEvent) MarshalJSON() ([]byte, error) {
	result, err := json.Marshal(self.Event)
	if err != nil {
		return nil, err
	}

	ret := []byte(fmt.Sprintf(`{"fd":%d,"t":%d,%s`, self.Fd, self.T, result[1:]))
	return ret, nil
}

type Event interface {
	EventType() string
}

type EventData struct {
	Type   string `json:"type"`
	Text   string `json:"raw"`
	Offset int    `json:"offset"`
}

func (self *EventData) EventType() string {
	return self.Type
}

type SerializedEvent struct {
	Inner Event
}

func (self *SerializedEvent) UnmarshalJSON(data []byte) error {
	eventData := new(EventData)
	err := json.Unmarshal(data, &eventData)
	if err != nil {
		return fmt.Errorf("json.Unmarshal(data, &eventData): %s", err)
	}
	switch eventData.Type {
	case "cursor":
		self.Inner = NewCursorEvent(eventData.Text, eventData.Offset)
	case "device":
		self.Inner = NewDeviceEvent(eventData.Text, eventData.Offset)
	case "display":
		self.Inner = NewDisplayEvent(eventData.Text, eventData.Offset)
	case "eof":
		self.Inner = NewEofEvent(eventData.Text, eventData.Offset)
	case "erase":
		self.Inner = NewEraseEvent(eventData.Text, eventData.Offset)
	case "fold":
		self.Inner = NewFoldEvent(eventData.Text, eventData.Offset)
	case "font":
		self.Inner = NewFontEvent(eventData.Text, eventData.Offset)
	case "not-implemented":
		self.Inner = NewNotImplementedEvent(eventData.Text, eventData.Offset)
	case "print":
		self.Inner = NewPrintEvent(eventData.Text, eventData.Offset)
	case "scroll":
		self.Inner = NewScrollEvent(eventData.Text, eventData.Offset)
	case "tab":
		self.Inner = NewTabEvent(eventData.Text, eventData.Offset)
	case "text":
		self.Inner = NewTextEvent(eventData.Text, eventData.Offset)
	default:
		return fmt.Errorf("unknown event type encountered: %s", eventData.Type)
	}
	return nil
}

func (self *SerializedEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.Inner)
}

type eventArgs struct {
	numeric []int
	str     string
}

func (e *eventArgs) n(index int) int {
	if index >= 0 && index < len(e.numeric) {
		return e.numeric[index]
	} else {
		return 0
	}
}

func (e *eventArgs) s() string {
	return e.str
}

func parseEventArgs(text string) *eventArgs {
	args := &eventArgs{}
	var scan func(input string)
	scan = func(input string) {
		if len(input) == 0 {
			return
		}
		switch input[0] {
		case ';':
			scan(input[1:])
		case '"':
			val := ""
			n, _ := fmt.Sscanf(input, "%q", &val)
			if n == 1 {
				args.str = val
				scan(input[len(val):])
			}
		default:
			val := 0
			n, _ := fmt.Sscanf(input, "%d", &val)
			if n == 1 {
				args.numeric = append(args.numeric, val)
			}
			if next := strings.Index(input, ";"); next != -1 {
				scan(input[next:])
			}
		}
	}

	if len(text) < 3 {
		return args
	}
	scan(text[2:])
	return args
}

type NotImplementedEvent struct {
	EventData
}

func NewNotImplementedEvent(text string, offset int) Event {
	return &NotImplementedEvent{
		EventData: EventData{"not-implemented", text, offset},
	}
}

type TextEvent struct {
	EventData
}

func NewTextEvent(text string, offset int) Event {
	return &TextEvent{
		EventData: EventData{"text", text, offset},
	}
}

type CursorAction int

const (
	CursorSave CursorAction = 1 + iota
	CursorRestore
	CursorSaveWithAttributes
	CursorRestoreWithAttributes
	CursorHome
	CursorForce
	CursorUp
	CursorDown
	CursorForward
	CursorBackward

	CursorBell
	CursorBackspace
	CursorCarriageReturn
	CursorFormFeed
	CursorHorizontalTab
	CursorLineFeed
	CursorVerticalTab
)

func (a CursorAction) String() string {
	switch a {
	case CursorBackspace:
		return "backspace"
	case CursorCarriageReturn:
		return "carriage-return"
	case CursorFormFeed:
		return "form-feed"
	case CursorHorizontalTab:
		return "horizontal-tab"
	case CursorLineFeed:
		return "line-feed"
	case CursorVerticalTab:
		return "vertical-tab"
	case CursorSave:
		return "save"
	case CursorRestore:
		return "restore"
	case CursorSaveWithAttributes:
		return "save-with-attributes"
	case CursorRestoreWithAttributes:
		return "restore-with-attributes"
	case CursorHome:
		return "home"
	case CursorForce:
		return "force"
	case CursorUp:
		return "up"
	case CursorDown:
		return "down"
	case CursorForward:
		return "forward"
	case CursorBackward:
		return "backward"
	default:
		return "unknown"
	}
}

func (a CursorAction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

type CursorEvent struct {
	EventData
	Action CursorAction `json:"action"`
	Count  int          `json:"count"`
	Row    int          `json:"row"`
	Column int          `json:"col"`
}

func NewCursorEvent(text string, offset int) Event {
	event := &CursorEvent{
		EventData: EventData{"cursor", text, offset},
	}
	args := parseEventArgs(text)

	switch text[len(text)-1] {
	case '\a':
		event.Action = CursorBell
	case '\b':
		event.Action = CursorBackspace
	case '\f':
		event.Action = CursorFormFeed
	case '\n':
		event.Action = CursorLineFeed
	case '\r':
		event.Action = CursorCarriageReturn
	case '\t':
		event.Action = CursorHorizontalTab
	case '\v':
		event.Action = CursorVerticalTab
	case 's':
		event.Action = CursorSave
	case 'u':
		event.Action = CursorRestore
	case '7':
		event.Action = CursorSaveWithAttributes
	case '8':
		event.Action = CursorRestoreWithAttributes
	case 'f':
		event.Action = CursorForce
		event.Row = args.n(0)
		event.Column = args.n(1)
	case 'H':
		event.Action = CursorHome
		event.Row = args.n(0)
		event.Column = args.n(1)
	case 'A':
		event.Action = CursorUp
		event.Count = args.n(0)
	case 'B':
		event.Action = CursorDown
		event.Count = args.n(0)
	case 'C':
		event.Action = CursorForward
		event.Count = args.n(0)
	case 'D':
		event.Action = CursorBackward
		event.Count = args.n(0)
	}

	return event
}

type FoldAction int

const (
	FoldOpen FoldAction = 1 + iota
	FoldClose
)

func (self FoldAction) String() string {
	switch self {
	case FoldOpen:
		return "open"
	case FoldClose:
		return "close"
	default:
		return "unknown"
	}
}

func (self FoldAction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + self.String() + `"`), nil
}

type FoldEvent struct {
	EventData

	Action FoldAction `json:"action"`
	Title  string     `json:"title"`
}

func NewFoldEvent(text string, offset int) Event {
	event := &FoldEvent{
		EventData: EventData{"fold", text, offset},
	}

	args := parseEventArgs(text)

	event.Title = args.s()
	if event.Title == "" {
		event.Action = FoldClose
	} else {
		event.Action = FoldOpen
	}

	return event
}

type FontEvent struct {
	EventData

	Font int `json:"font"`
}

func NewFontEvent(text string, offset int) Event {
	return &FontEvent{
		EventData: EventData{"font", text, offset},
	}
}

type DeviceAction int

const (
	DeviceActionReset DeviceAction = 1 + iota
	DeviceEnableLineWrap
	DeviceDisableLineWrap
	DeviceQueryCode
	DeviceCode
	DeviceQueryStatus
	DeviceStatus
	DeviceQueryCursor
	DeviceCursor
)

func (d DeviceAction) String() string {
	switch d {
	case DeviceActionReset:
		return "reset"
	case DeviceEnableLineWrap:
		return "enable-line-wrap"
	case DeviceDisableLineWrap:
		return "disable-line-wrap"
	case DeviceQueryCode:
		return "query-code"
	case DeviceCode:
		return "code"
	case DeviceQueryStatus:
		return "query-status"
	case DeviceStatus:
		return "status"
	case DeviceQueryCursor:
		return "query-cursor"
	case DeviceCursor:
		return "cursor"
	default:
		return "unknown"
	}
}

func (d DeviceAction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

type DeviceEvent struct {
	EventData

	LineWrap bool `json:"lineWrap"`
	StatusOK bool `json:"statusOk"`
	Code     int  `json:"code"`

	Row    int `json:"row"`
	Column int `json:"col"`
}

func NewDeviceEvent(text string, offset int) Event {
	return &DeviceEvent{
		EventData: EventData{"device", text, offset},
	}
}

type EraseTarget int

const (
	EraseToEOL EraseTarget = 1 + iota
	EraseToBOL
	EraseLine
	EraseDown
	EraseUp
	EraseScreen
)

func (t EraseTarget) String() string {
	switch t {
	case EraseToEOL:
		return "eol"
	case EraseToBOL:
		return "bol"
	case EraseLine:
		return "line"
	case EraseDown:
		return "down"
	case EraseUp:
		return "up"
	case EraseScreen:
		return "screen"
	default:
		return "unknown"
	}
}

func (t EraseTarget) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

type EraseEvent struct {
	EventData

	Target EraseTarget `json:"target"`
}

func NewEraseEvent(text string, offset int) Event {
	return &EraseEvent{
		EventData: EventData{"erase", text, offset},
	}
}

type DisplayAttribute int

const (
	DisplayNone       = -1
	DisplayReset      = 0
	DisplayBright     = 1
	DisplayDim        = 2
	DisplayUnderscore = 4
	DisplayBlink      = 5
	DisplayReverse    = 7
	DisplayHidden     = 8
)

func (a DisplayAttribute) String() string {
	switch a {
	case DisplayNone:
		return "none"
	case DisplayReset:
		return "reset"
	case DisplayBright:
		return "bright"
	case DisplayDim:
		return "dim"
	case DisplayUnderscore:
		return "underscore"
	case DisplayBlink:
		return "blink"
	case DisplayReverse:
		return "reverse"
	case DisplayHidden:
		return "hidden"
	default:
		return "unknown"
	}
}

func (a DisplayAttribute) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

type DisplayEvent struct {
	EventData

	Attribute DisplayAttribute `json:"attr"`

	// Foreground is the foreground color to use.  If it is 0, the
	// foreground color doesn't change.  If it is greater than
	// zero it is the index into the color palette + 1
	Foreground int `json:"fg"`
	// Background is the background color to use.  If it is 0, the
	// background color doesn't change.  If it is greater than
	// zero it is the index into the color palette + 1
	Background int `json:"bg"`
}

func NewDisplayEvent(text string, offset int) Event {
	event := &DisplayEvent{
		EventData: EventData{"display", text, offset},
		Attribute: DisplayNone,
	}

	args := parseEventArgs(text)
	event.evalArgs(args)

	return event
}

func (self *DisplayEvent) evalArgs(args *eventArgs) {
	if self.isExtendedColor(args) {
		self.evalExtendedColor(args)
		return
	}

	for _, arg := range args.numeric {
		if arg >= 0 && arg <= 8 {
			self.Attribute = DisplayAttribute(arg)
		} else if arg >= 30 && arg <= 37 {
			self.Foreground = arg - 29
		} else if arg >= 40 && arg <= 49 {
			self.Background = arg - 39
		}
	}
}

func (self *DisplayEvent) isExtendedColor(args *eventArgs) bool {
	if len(args.numeric) < 3 {
		return false
	}
	prefix := fmt.Sprintf("%d %d", args.numeric[0], args.numeric[1])
	switch prefix {
	case "38 5":
		return true
	case "48 5":
		return true
	default:
		return false
	}
}

func (self *DisplayEvent) evalExtendedColor(args *eventArgs) {
	// CSI [ 38 5 $COLOR -> set foreground to $COLOR+1
	// CSI [ 48 5 $COLOR -> set background to $COLOR+1
	// where $COLOR in {0..255}
	dst := (*int)(nil)
	switch args.numeric[0] {
	case 38:
		dst = &self.Foreground
	case 48:
		dst = &self.Background
	default:
		return
	}

	if args.numeric[1] != 5 {
		return
	}

	*dst = args.numeric[2] + 1
}

type TabAction int

const (
	TabSetCurrent TabAction = 1 + iota
	TabClearCurrent
	TabClearAll
)

func (t TabAction) String() string {
	switch t {
	case TabSetCurrent:
		return "set-current"
	case TabClearCurrent:
		return "clear-current"
	case TabClearAll:
		return "clear-all"
	default:
		return "unknown"
	}
}

func (t TabAction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

type TabEvent struct {
	EventData
	Action TabAction `json:"action"`
}

func NewTabEvent(text string, offset int) Event {
	return &TabEvent{
		EventData: EventData{"tab", text, offset},
	}
}

type PrintAction int

const (
	PrintScreen PrintAction = 1 + iota
	PrintLine
	PrintStartLog
	PrintStopLog
)

func (a PrintAction) String() string {
	switch a {
	case PrintScreen:
		return "screen"
	case PrintLine:
		return "line"
	case PrintStartLog:
		return "start"
	case PrintStopLog:
		return "stop"
	default:
		return "unknown"
	}
}

func (a PrintAction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

type PrintEvent struct {
	EventData

	Action PrintAction `json:"action"`
}

func NewPrintEvent(text string, offset int) Event {
	return &PrintEvent{
		EventData: EventData{"print", text, offset},
	}
}

type ScrollAction int

const (
	ScrollDown ScrollAction = 1 + iota
	ScrollUp
	ScrollScreen
	ScrollRegion
)

func (a ScrollAction) String() string {
	switch a {
	case ScrollDown:
		return "down"
	case ScrollUp:
		return "up"
	case ScrollScreen:
		return "screen"
	case ScrollRegion:
		return "region"
	default:
		return "unknown"
	}
}

func (a ScrollAction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

type ScrollEvent struct {
	EventData

	Action ScrollAction `json:"action"`
	Start  int          `json:"start"`
	End    int          `json:"end"`
}

func NewScrollEvent(text string, offset int) Event {
	return &ScrollEvent{
		EventData: EventData{"scroll", text, offset},
	}
}

type EofEvent struct {
	EventData
}

func (self EofEvent) EventType() string {
	return "eof"
}

func NewEofEvent(text string, offset int) Event {
	return &EofEvent{
		EventData: EventData{"eof", text, offset},
	}
}

var EOF = NewEofEvent("", 0)
