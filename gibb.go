/*
Package gibb implements a method of broadcasting messages to receivers with
the following guarantees.

No messages are dropped. Each receiver will always get every message that
is broadcasted.

Slow receivers never block delivery to other receivers that are also
listening.

All messages are delivered in order.

Potentially unbounded memory use if consumers cannot keep up with producers.
Receivers that fall out of scope will be GC'd, but if a receiver is present and
not read from, its memory use will grow as values are written to its
broadcaster.
*/
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

// New creates a new broadcaster with the necessary internal
// structure. The uninitialized broadcaster is unsuitable to be listened or
// written to.
func New() *Broadcaster {
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

func (r *Receiver) readCancel(c <-chan struct{}) (msg message, ok bool) {
	select {
	case <-c:
		return message{}, false

	default:
	}

	select {
	case <-c:
		return message{}, false

	case msg := <-r.c:
		r.c <- msg
		return msg, true
	}
}

func (r *Receiver) read() message {
	msg := <-r.c
	r.c <- msg

	return msg
}

func (r *Receiver) swap(m message) {
	r.c = m.c
}

func always(cf func(<-chan struct{}) CancelledResult) interface{} {
	return cf(make(chan struct{})).Val
}

// Read a single value that has been broadcast. Blocks until messages have been
// written.
func (r *Receiver) Read() interface{} {
	return always(r.ReadCancel)
}

// Peek reads a single value that has been broadcast and then places it back
// onto the receiver. Blocks until messages have been written.
func (r *Receiver) Peek() interface{} {
	return always(r.PeekCancel)
}

// MustReadVal reads from a Receiver until it gets a value that is assignable
// to the given pointer. If a pointer is not supplied, this method will never
// return.
func (r *Receiver) MustReadVal(v interface{}) {
	for !r.ReadVal(v) {
		r.Read()
	}
}

// ReadVal reads a value from the Reciver and attempts to write it into the
// given pointer. If the read value can not be assigned to the given interface
// for any reason, false will be returned and the value will be placed back onto
// the receiver.
func (r *Receiver) ReadVal(v interface{}) bool {
	r.mx.Lock()
	defer r.mx.Unlock()

	msg := r.read()
	ok := fill.Fill(v, msg.v) == nil
	if ok {
		r.swap(msg)
	}

	return ok
}

// ReadChan locks the receiver and writes any broadcasted messages to the output
// channel. When you are ready to stop receiving messages, close the second
// signaling channel.
func (r *Receiver) ReadChan() (<-chan interface{}, chan<- struct{}) {
	vc := make(chan interface{})
	cc := make(chan struct{})

	go func() {
		defer close(vc)

		for {
			result := r.ReadCancel(cc)
			if !result.Okay {
				return
			}

			vc <- result.Val
		}
	}()

	return vc, cc
}

type CancelledResult struct {
	Okay bool
	Val  interface{}
}

func (r *Receiver) ReadCancel(c <-chan struct{}) CancelledResult {
	r.mx.Lock()
	defer r.mx.Unlock()

	msg, ok := r.readCancel(c)
	if ok {
		r.swap(msg)
	}

	return CancelledResult{ok, msg.v}
}

func (r *Receiver) PeekCancel(c <-chan struct{}) CancelledResult {
	r.mx.Lock()
	defer r.mx.Unlock()

	msg, ok := r.readCancel(c)
	return CancelledResult{ok, msg.v}
}

type CancelledFill struct {
	DidRead bool
	DidFill bool
}

func (cf CancelledFill) Good() bool {
	return cf.DidRead && cf.DidFill
}

func (r *Receiver) ReadValCancel(v interface{}, c <-chan struct{}) CancelledFill {
	r.mx.Lock()
	defer r.mx.Unlock()

	var result CancelledFill

	msg, ok := r.readCancel(c)
	result.DidRead = ok

	if !ok {
		return result
	}

	ok = fill.Fill(v, msg.v) == nil
	result.DidFill = ok

	if !ok {
		return result
	}

	r.swap(msg)
	return result
}

func (r *Receiver) MustReadValCancel(v interface{}, c <-chan struct{}) bool {
	result := r.ReadValCancel(v, c)

	for !result.Good() {
		if !result.DidRead {
			return false
		}

		r.Read()
		result = r.ReadValCancel(v, c)
	}

	return true
}
