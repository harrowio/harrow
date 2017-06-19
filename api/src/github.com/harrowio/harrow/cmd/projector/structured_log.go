package projector

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"strings"
	"time"
)

type Clock interface {
	Now() time.Time
}

type StructuredLog struct {
	clock   Clock
	size    int
	current int
	entries []StructuredLogEntry
}

type StructuredLogEntry struct {
	Level      string
	OccurredOn time.Time
	Fields     map[string][]string
}

func (self *StructuredLogEntry) Parse(level string, now time.Time, text string) {
	self.Level = level
	self.OccurredOn = now

	self.Fields = map[string][]string{}

	fieldDefinitions := strings.Fields(text)
	for _, definition := range fieldDefinitions {
		kvSeparatorAt := strings.Index(definition, "=")
		key, value := "", ""
		if kvSeparatorAt == -1 {
			key = "text"
			value = definition
		} else {
			key = definition[:kvSeparatorAt]
			if len(definition) > kvSeparatorAt+1 {
				value = definition[kvSeparatorAt+1:]
			}
		}

		self.Fields[key] = append(self.Fields[key], value)
	}
}

func (self *StructuredLogEntry) String() string {
	timeField := fmt.Sprintf("%%-%ds", len("2016-07-30T13:20:14.334347899Z"))
	out := bytes.NewBufferString(
		fmt.Sprintf(timeField+" %-10s",
			self.OccurredOn.Format(time.RFC3339Nano),
			strings.ToUpper(self.Level),
		),
	)
	for fieldName, fieldValues := range self.Fields {
		if fieldName == "text" {
			fmt.Fprintf(out, " %s", strings.Join(fieldValues, " "))
			continue
		}

		for _, value := range fieldValues {
			fmt.Fprintf(out, " %s=%s", fieldName, value)
		}
	}

	return out.String()
}

func NewStructuredLog(size int, clock Clock) *StructuredLog {
	return &StructuredLog{
		size:    size,
		entries: make([]StructuredLogEntry, size),
		current: 0,
		clock:   clock,
	}
}

func (self *StructuredLog) RenderText(w io.Writer) {
	self.Each(func(entry *StructuredLogEntry) {
		fmt.Fprintf(w, "%s\n", entry)
	})
}

func (self *StructuredLog) RenderHTML(w io.Writer) {
	fmt.Fprintf(w, `<table><tbody>`)
	self.Each(func(entry *StructuredLogEntry) {
		fmt.Fprintf(w, `<tr>
<td style="padding-right: 1ex">%s</td>
<td style="text-transform: uppercase">%s</td>`,
			html.EscapeString(entry.OccurredOn.Format(time.RFC3339Nano)),
			html.EscapeString(entry.Level),
		)
		for fieldName, fieldValues := range entry.Fields {
			for _, value := range fieldValues {
				if fieldName != "text" {
					fmt.Fprintf(w, `<td>%s</td>`,
						html.EscapeString(fieldName),
					)
				}

				fmt.Fprintf(w, `<td>%s</td>`,
					html.EscapeString(value),
				)
			}
		}
		fmt.Fprintf(w, `</tr>`)
	})
	fmt.Fprintf(w, `</tbody></table>`)
}

func (self *StructuredLog) Each(do func(entry *StructuredLogEntry)) *StructuredLog {
	for i := 0; i < self.size; i++ {
		entry := &self.entries[(self.current+i)%self.size]
		if !entry.OccurredOn.IsZero() {
			do(entry)
		}
	}

	return self
}

func (self *StructuredLog) Fatal(thing interface{}) {
	panic(thing)
}

func (self *StructuredLog) Infof(format string, args ...interface{}) error {
	text := fmt.Sprintf(format, args...)
	self.LogMessage("info", text)
	return nil
}

func (self *StructuredLog) Errf(format string, args ...interface{}) error {
	text := fmt.Sprintf(format, args...)
	self.LogMessage("error", text)
	return nil
}

func (self *StructuredLog) LogMessage(level string, text string) *StructuredLog {
	entry := self.nextEntry()
	entry.Parse(level, self.clock.Now(), text)
	return self
}

func (self *StructuredLog) nextEntry() *StructuredLogEntry {
	entry := &self.entries[self.current%self.size]
	self.current++
	return entry
}
