package broadcast

import "time"

type MemoryMessage struct {
	DoUUID          func() string
	DoTable         func() string
	DoType          func() BroadcastMessageType
	DoAcknowledge   func() error
	DoRejectOnce    func() error
	DoRejectForever func() error
	DoRequeueAfter  func(d time.Duration) error

	Acknowledged    bool
	RejectedOnce    bool
	RejectedForever bool
	RequeuedAfter   time.Duration
}

func NewMemoryMessage(table, uuid string, typ BroadcastMessageType) *MemoryMessage {
	msg := &MemoryMessage{
		DoTable: func() string { return table },
		DoUUID:  func() string { return uuid },
		DoType:  func() BroadcastMessageType { return typ },
	}

	msg.DoAcknowledge = func() error { msg.Acknowledged = true; return nil }
	msg.DoRejectOnce = func() error { msg.RejectedOnce = true; return nil }
	msg.DoRejectForever = func() error { msg.RejectedForever = true; return nil }
	msg.DoRequeueAfter = func(d time.Duration) error { msg.RequeuedAfter = d; return nil }

	return msg
}

func (self *MemoryMessage) Table() string              { return self.DoTable() }
func (self *MemoryMessage) UUID() string               { return self.DoUUID() }
func (self *MemoryMessage) Type() BroadcastMessageType { return self.DoType() }
func (self *MemoryMessage) Acknowledge() error         { return self.DoAcknowledge() }
func (self *MemoryMessage) RejectOnce() error          { return self.DoRejectOnce() }
func (self *MemoryMessage) RejectForever() error       { return self.DoRejectForever() }
func (self *MemoryMessage) RequeueAfter(d time.Duration) error {
	return self.DoRequeueAfter(d)
}

func NewMemoryMessageCreate(table, uuid string) *MemoryMessage {
	return NewMemoryMessage(table, uuid, Create)
}

func NewMemoryMessageChange(table, uuid string) *MemoryMessage {
	return NewMemoryMessage(table, uuid, Change)
}
