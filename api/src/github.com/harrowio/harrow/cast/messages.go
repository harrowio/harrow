package cast

import "strings"

type MessageType string
type Payload map[string][]string

func (self Payload) Get(key string) string {
	values, found := self[key]
	if !found {
		return ""
	}

	return values[0]
}

func (self Payload) Set(key, value string) Payload {
	self[key] = []string{value}
	return self
}

const (
	ChildExited MessageType = "child_exited"
	FatalError  MessageType = "fatal_error"
	Heartbeat   MessageType = "heartbeat"
	Event       MessageType = "event"
)

func (self MessageType) String() string { return string(self) }

// ControlMessage captures the structure of a message sent on the
// control channel of the cast command.
type ControlMessage struct {
	// Type indicates the kind of event that has occurred
	Type MessageType `json:"type"`

	// ExitStatus is the exit status of the command
	ExitStatus int `json:"exitStatus"`

	// Payload that provides additional information about the
	// message
	Payload Payload `json:"payload"`
}

func NewStatusLogEntry(entryType, subject string) *ControlMessage {
	return &ControlMessage{
		Type: Event,
		Payload: Payload{
			"event":   []string{"status"},
			"type":    []string{entryType},
			"subject": []string{strings.TrimSpace(subject)},
			"body":    []string{""},
		},
	}
}

func NewOutputMessage(tag string, data []byte) *ControlMessage {
	return &ControlMessage{
		Type: Event,
		Payload: Payload{
			"type":    []string{"output"},
			"text":    []string{string(data)},
			"channel": []string{tag},
		},
	}
}
