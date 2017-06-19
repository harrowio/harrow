package cast

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrSinkTimedOut = errors.New("sink: recv timed out")
)

// Sink is a simple TCP server capturing any messages it receives and
// closing connections after a given amount of time.
type Sink struct {
	listenOn string          // connection information for listening
	received *bytes.Buffer   // received bytes
	requests *sync.WaitGroup // pending requests
	errors   []error         // any errors encountered during processing
	timeout  time.Duration   // time after which to terminate the connection
}

// NewSink creates a Sink listening on listenOn.
func NewSink(listenOn string) *Sink {
	return &Sink{
		listenOn: listenOn,
		received: new(bytes.Buffer),
		requests: &sync.WaitGroup{},
		timeout:  100 * time.Millisecond,
	}
}

// Start starts listening for and accepting connections.
func (self *Sink) Start() error {
	ln, err := net.Listen("tcp", self.listenOn)
	if err != nil {
		return err
	}

	go self.acceptLoop(ln)

	return nil
}

func (self *Sink) acceptLoop(ln net.Listener) {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		done := make(chan bool, 1)

		self.requests.Add(1)
		go func(conn net.Conn, received *bytes.Buffer) {
			defer conn.Close()

			if _, err := io.Copy(received, conn); err != nil {
				self.recordError(conn.RemoteAddr(), err)
			}

			done <- true
		}(conn, self.received)

		timeout := time.After(self.timeout)
		select {
		case <-timeout:
			self.recordError(conn.RemoteAddr(), ErrSinkTimedOut)
			conn.Close()
		case <-done:
		}

		self.requests.Done()
	}
}

// Errors returns any errors the sink encountered while operating.
func (self *Sink) Errors() []error {
	return self.errors
}

func (self *Sink) recordError(addr net.Addr, err error) {
	self.errors = append(self.errors, fmt.Errorf("sink: copy from %s: %s", addr, err))
}

// Reset clears the sink's buffer of received text.
func (self *Sink) Reset() {
	self.requests.Wait()
	self.received.Reset()
}

// Recv waits for all requests against this sink to finish and then
// returned the buffer of received text.
func (self *Sink) Recv() *bytes.Buffer {
	self.requests.Wait()
	return self.received
}
