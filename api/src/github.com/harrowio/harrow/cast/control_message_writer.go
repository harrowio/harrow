package cast

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type ControlMessageWriter struct {
	tag  string
	out  *bufio.Writer
	dest io.Writer
}

func NewControlMessageWriter(tag string, out io.Writer) *ControlMessageWriter {
	return &ControlMessageWriter{
		dest: out,
		out:  bufio.NewWriter(out),
		tag:  tag,
	}
}

func (self *ControlMessageWriter) Write(p []byte) (n int, err error) {
	msg := NewOutputMessage(self.tag, p)
	data, err := json.Marshal(msg)

	if err != nil {
		return 0, err
	}

	if _, err := fmt.Fprintf(self.out, "%s\n", data); err != nil {
		return 0, err
	}

	if err := self.out.Flush(); err != nil {
		return 0, err
	}

	self.out.Reset(self.dest)

	return len(p), nil
}
