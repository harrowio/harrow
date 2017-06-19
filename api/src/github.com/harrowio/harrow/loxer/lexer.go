package loxer

type eventConstructor func(text string, offset int) Event

var asciiCtrl = map[byte]eventConstructor{
	'\a': NewDeviceEvent,
	'\b': NewCursorEvent,
	'\f': NewCursorEvent,
	'\n': NewCursorEvent,
	'\r': NewCursorEvent,
	'\t': NewCursorEvent,
	'\v': NewCursorEvent,
}

var oscSeqTerminators = map[byte]eventConstructor{
	'\a': NewFoldEvent,
}

var escapeSeq = map[byte]eventConstructor{
	'(': NewDeviceEvent,
	')': NewDeviceEvent,
	'c': NewDeviceEvent,
	'7': NewCursorEvent,
	'8': NewCursorEvent,
	'D': NewScrollEvent,
	'H': NewTabEvent,
}

var ctrlSeqTerminators = map[byte]eventConstructor{
	'A': NewCursorEvent,
	'B': NewCursorEvent,
	'C': NewCursorEvent,
	'D': NewCursorEvent,
	'f': NewCursorEvent,
	'R': NewCursorEvent,
	's': NewCursorEvent,
	'u': NewCursorEvent,
	'c': NewDeviceEvent,
	'n': NewDeviceEvent,
	'm': NewDisplayEvent,
	'K': NewEraseEvent,
	'J': NewEraseEvent,
	'p': NewNotImplementedEvent,
	'h': NewDeviceEvent,
	'l': NewDeviceEvent,
	'i': NewPrintEvent,
	'M': NewScrollEvent,
	'r': NewScrollEvent,
	'g': NewTabEvent,
}

type lexer struct {
	offset   int
	state    stateFn
	tokbuf   []byte
	listener func(Event)
}

type stateFn func(b byte) stateFn

func NewLexer(h func(Event)) *lexer {
	return newLexer(h)
}

func (l *lexer) Write(p []byte) (n int, err error) {
	l.feed(p)
	return len(p), nil
}

func newLexer(listener func(Event)) *lexer {
	l := &lexer{
		listener: listener,
	}
	l.state = l.stateRawText
	return l
}

func (l *lexer) feed(p []byte) {
	for _, b := range p {
		l.feedByte(b)
	}
}

func (l *lexer) feedByte(b byte) {
	l.state = l.state(b)
}

func (l *lexer) stateRawText(b byte) stateFn {
	if b == '\033' {
		l.emitEventWithCurBuffer(NewTextEvent)
		l.appendTokenBuffer('\033')
		return l.stateEscapeSeq
	} else if asciiEventType, found := asciiCtrl[b]; found {
		l.emitEventWithCurBuffer(NewTextEvent)
		l.appendTokenBuffer(b)
		l.emitEventWithCurBuffer(asciiEventType)
		return l.stateRawText
	} else {
		l.appendTokenBuffer(b)
		return l.stateRawText
	}
}

func (l *lexer) stateEscapeSeq(b byte) stateFn {
	if seqName, found := escapeSeq[b]; found {
		l.emitEventWithCurBuffer(seqName)
		return l.stateRawText
	} else if b == '[' {
		l.appendTokenBuffer(b)
		return l.stateCtrlSeq
	} else if b == ']' {
		l.appendTokenBuffer(b)
		return l.stateOperatingSystemCommand
	} else {
		return l.stateRawText
	}
}

func (l *lexer) stateOperatingSystemCommand(b byte) stateFn {
	l.appendTokenBuffer(b)
	if oscSeqGroup, found := oscSeqTerminators[b]; found {
		l.emitEventWithCurBuffer(oscSeqGroup)
		return l.stateRawText
	} else {
		return l.stateOperatingSystemCommand
	}
}

func (l *lexer) stateCtrlSeq(b byte) stateFn {
	l.appendTokenBuffer(b)
	if ctrlSeqGroup, found := ctrlSeqTerminators[b]; found {
		l.emitEventWithCurBuffer(ctrlSeqGroup)
		return l.stateRawText
	} else {
		return l.stateCtrlSeq
	}
}

func (l *lexer) appendTokenBuffer(b byte) *lexer {
	l.tokbuf = append(l.tokbuf, b)
	return l
}

func (l *lexer) clear() {
	l.tokbuf = nil
}

func (l *lexer) tokenBuffer() []byte {
	return l.tokbuf
}

func (l *lexer) Close() error {
	if len(l.tokenBuffer()) > 0 {
		l.emitEventWithCurBuffer(NewTextEvent)
	}

	return nil
}

func (l *lexer) emitEventWithCurBuffer(newEvent eventConstructor) {
	tok := l.tokenBuffer()
	if len(tok) == 0 {
		return
	}

	event := newEvent(string(tok), l.offset)
	l.offset = l.offset + len(tok)
	l.listener(event)
	l.clear()
}
