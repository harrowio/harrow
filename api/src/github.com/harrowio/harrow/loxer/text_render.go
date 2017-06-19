package loxer

import (
	"bytes"
	"fmt"
)

type TextRenderer struct {
	result *bytes.Buffer
}

func NewTextRenderer() *TextRenderer {
	return &TextRenderer{
		result: bytes.NewBufferString(""),
	}
}

func (self *TextRenderer) Handle(event Event) {
	switch e := event.(type) {
	case *TextEvent:
		fmt.Fprintf(self.result, "%s", e.Text)
	case *CursorEvent:
		if e.Action == CursorHorizontalTab {
			fmt.Fprintf(self.result, "  ")
		} else {
			fmt.Fprintf(self.result, "%s", e.Text)
		}
	}
}

func (self *TextRenderer) String() string {
	return self.result.String()
	// Fd= 2 == ERR
}
