package cast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// ControlMessageParser implements the io.Writer interface by scanning
// bytes that are written to it for valid JSON objects, decoding those
// into ControlMessages and then sending those messages onto a
// channel.
//
// It closes the channel when it receives a `ChildExited` message.
type ControlMessageParser struct {
	results        chan *ControlMessage
	currentMessage *bytes.Buffer
}

// NewControlMessageParser constructs a new control message parser
// which reports new messages on the channel results.
func NewControlMessageParser(results chan *ControlMessage) *ControlMessageParser {
	return &ControlMessageParser{
		results:        results,
		currentMessage: bytes.NewBufferString(""),
	}
}

func (self *ControlMessageParser) Write(p []byte) (int, error) {
	message := &ControlMessage{}

	fmt.Fprintf(self.currentMessage, "%s", p)

	dec := json.NewDecoder(self.currentMessage)
	err := dec.Decode(message)
	for err == nil {
		self.results <- message
		if message.Type == ChildExited {
			close(self.results)
			self.results = nil
		}

		message = &ControlMessage{}
		err = dec.Decode(message)
	}
	rest := bytes.NewBufferString("")
	io.Copy(rest, dec.Buffered())
	self.currentMessage = rest

	return len(p), nil
}
