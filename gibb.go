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
	"fmt"
	"sync"
	"time"

	"github.com/dagoof/fill"
)

func Timeout(d time.Duration) <-chan struct{} {
	stop := make(chan struct{})

	go func() {
		<-time.After(d)
		close(stop)
	}()

	return stop
}

func Val(dst, src interface{}) bool {
	return fill.Fill(dst, src) == nil
}

// message contains a snapshot of a value.
type message struct {
	v interface{}
	c chan message
}

type Maybe struct {
	V      interface{}
	Exists bool
}

func (m Maybe) Describe() string {
	if !m.Exists {
		return "nothing"
	}

	return fmt.Sprintf("just %v", m.V)
}

type MaybeVal struct {
	Maybe
	Valid bool
}

func (m MaybeVal) Describe() string {
	if !m.Valid {
		return "invalid"
	}

	return m.Maybe.Describe()
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

func (r *Receiver) peek() message {
	msg := <-r.c
	r.c <- msg
	return msg
}

func (r *Receiver) swap(m message) {
	r.c = m.c
}

func (r *Receiver) read() message {
	msg := r.peek()
	r.swap(msg)

	return msg
}

// Read a single value that has been broadcast. Blocks until messages have been
// written.
func (r *Receiver) Read() interface{} {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.read().v
}

func (r *Receiver) ReadCancel(stop <-chan struct{}) Maybe {
	r.mx.Lock()
	defer r.mx.Unlock()

	select {
	case msg := <-r.c:
		r.c <- msg
		r.swap(msg)
		return Maybe{msg.v, true}
	case <-stop:
	}

	return Maybe{}
}

func (r *Receiver) ReadTimeout(d time.Duration) Maybe {
	return r.ReadCancel(Timeout(d))
}

// Peek reads a single value that has been broadcast and then places it back
// onto the receiver. Blocks until messages have been written.
func (r *Receiver) Peek() interface{} {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.peek().v
}

func (r *Receiver) PeekCancel(stop <-chan struct{}) Maybe {
	r.mx.Lock()
	defer r.mx.Unlock()

	select {
	case msg := <-r.c:
		r.c <- msg
		return Maybe{msg.v, true}
	case <-stop:
	}

	return Maybe{}
}

// ReadVal reads a value from the Reciver and attempts to write it into the
// gven pointer. If the read value can not be assigned to the given interface
// for any reason, false will be returned and the value will be placed back onto
// the receiver.
func (r *Receiver) ReadVal(v interface{}) bool {
	r.mx.Lock()
	defer r.mx.Unlock()

	var (
		msg = r.peek()
		ok  = Val(v, msg.v)
	)

	if ok {
		r.swap(msg)
	}

	return ok
}

func (r *Receiver) ReadValCancel(stop <-chan struct{}, v interface{}) MaybeVal {
	r.mx.Lock()
	defer r.mx.Unlock()

	select {
	case msg := <-r.c:
		r.c <- msg
		ok := Val(v, msg.v)
		if ok {
			r.swap(msg)
		}

		return MaybeVal{Maybe{msg.v, true}, ok}
	case <-stop:
	}

	return MaybeVal{}
}

func (r *Receiver) ReadValTimeout(d time.Duration, v interface{}) MaybeVal {
	return r.ReadValCancel(Timeout(d), v)
}

// MustReadVal reads from a Receiver until it gets a value that is assignable
// to the given pointer. If a pointer is not supplied, this method will never
// return.
func (r *Receiver) MustReadVal(v interface{}) {
	r.mx.Lock()
	defer r.mx.Unlock()

	var (
		msg message
		ok  bool
	)

	for {
		msg = r.peek()
		r.swap(msg)
		ok = Val(v, msg.v)

		if ok {
			return
		}
	}
}

func (r *Receiver) MustReadValCancel(stop <-chan struct{}, v interface{}) bool {
	r.mx.Lock()
	defer r.mx.Unlock()

	var (
		msg message
		ok  bool
	)

	for {
		select {
		case msg = <-r.c:
			r.c <- msg
			r.swap(msg)

			ok = Val(v, msg.v)
			if ok {
				return true
			}
		case <-stop:
			return false
		}
	}

	return false
}
