package broadcast

type MemoryTransport struct {
	types     map[BroadcastMessageType]chan Message
	published map[BroadcastMessageType][]Message
	handler   func(*MemoryMessage)
}

func NewMemoryTransport() *MemoryTransport {
	return &MemoryTransport{
		types: map[BroadcastMessageType]chan Message{
			Create: make(chan Message, 1),
			Change: make(chan Message, 1),
		},
		published: map[BroadcastMessageType][]Message{},
		handler:   func(*MemoryMessage) {},
	}
}

func (self *MemoryTransport) Consume(typ BroadcastMessageType) (<-chan Message, error) {
	return self.types[typ], nil
}

func (self *MemoryTransport) Published(typ BroadcastMessageType) []Message {
	return self.published[typ]
}

func (self *MemoryTransport) Publish(typ, table, uuid string) error {
	t := BroadcastMessageType(typ)
	message := NewMemoryMessage(table, uuid, t)
	self.handler(message)
	self.published[t] = append(self.published[t], message)
	self.types[t] <- message
	return nil
}

func (self *MemoryTransport) Inspect(handler func(msg *MemoryMessage)) *MemoryTransport {
	self.handler = handler
	return self
}

func (self *MemoryTransport) Close() error {
	for _, publish := range self.types {
		close(publish)
	}
	return nil
}
