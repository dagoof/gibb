package gibb

import (
	"sync"

	"github.com/dagoof/fill"
)

// message contains a snapshot of a value.
type message struct {
	v interface{}
	c chan message
}

// Broadcaster provides Receivers that can be read from. Every value that is
// written to a broadcaster will be sent to any active Receivers. No messages
// are dropped, and are delivered in order. Receivers that fail to keep up with
// producers can oom your system because no messages are dropped.
type Broadcaster struct {
	mx sync.Mutex
	c  chan message
}

// Receiver can be Read from in various ways. All reads are concurrency-safe.
type Receiver struct {
	mx sync.Mutex
	c  chan message
}

// NewBroadcaster creates a new broadcaster with the necessary internal
// structure. The uninitialized broadcaster is unsuitable to be listened or
// written to.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		sync.Mutex{},
		make(chan message, 1),
	}
}

// Write a value to all listening receivers.
func (b *Broadcaster) Write(v interface{}) {
	c := make(chan message, 1)
	b.mx.Lock()

	b.c <- message{v, c}
	b.c = c

	b.mx.Unlock()
}

// Listen creates a receiver that can read written values.
func (b *Broadcaster) Listen() *Receiver {
	b.mx.Lock()
	defer b.mx.Unlock()

	return &Receiver{sync.Mutex{}, b.c}
}

// Read a single value that has been broadcast. Blocks until messages have been
// written.
func (r *Receiver) Read() interface{} {
	r.mx.Lock()
	defer r.mx.Unlock()

	msg := <-r.c
	r.c <- msg

	r.c = msg.c
	return msg.v
}

// ReadVal reads from a Receiver until it gets a value that is assignable to the
// given pointer. If a pointer is not supplied, this method will never return.
func (r *Receiver) ReadVal(v interface{}) {
	for {
		if fill.Fill(v, r.Read()) == nil {
			return
		}
	}
}

// ReadChan locks the receiver and writes any broadcasted messages to the output
// channel. When you are ready to stop receiving messages, close the second
// signaling channel.
func (r *Receiver) ReadChan() (<-chan interface{}, chan struct{}) {
	vc := make(chan interface{})
	cc := make(chan struct{})

	go func() {
		r.mx.Lock()
		defer r.mx.Unlock()
		defer close(vc)

		for {
			select {
			case <-cc:
				return
			case msg := <-r.c:
				r.c <- msg
				r.c = msg.c

				vc <- msg.v
			}
		}
	}()

	return vc, cc
}
