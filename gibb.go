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
)

// message contains a snapshot of a value.
type message[T any] struct {
	v T
	c chan message[T]
}

// Broadcaster provides Receivers that can be read from. Every value that is
// written to a broadcaster will be sent to any active Receivers. No messages
// are dropped, and are delivered in order. Receivers that fail to keep up with
// producers can oom your system because no messages are dropped.
type Broadcaster[T any] struct {
	mx sync.Mutex
	c  chan message[T]
}

// Receiver can be Read from in various ways. All reads are concurrency-safe.
type Receiver[T any] struct {
	mx sync.Mutex
	c  chan message[T]
}

// New creates a new broadcaster with the necessary internal
// structure. The uninitialized broadcaster is unsuitable to be listened or
// written to.
func New[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		sync.Mutex{},
		make(chan message[T], 1),
	}
}

// Write a value to all listening receivers.
func (b *Broadcaster[T]) Write(v T) {
	c := make(chan message[T], 1)
	b.mx.Lock()

	b.c <- message[T]{v, c}
	b.c = c

	b.mx.Unlock()
}

// Listen creates a receiver that can read written values.
func (b *Broadcaster[T]) Listen() *Receiver[T] {
	b.mx.Lock()
	defer b.mx.Unlock()

	return &Receiver[T]{sync.Mutex{}, b.c}
}

func (r *Receiver[T]) readCancel(c <-chan struct{}) (msg message[T], ok bool) {
	select {
	case <-c:
		return message[T]{}, false

	default:
	}

	select {
	case <-c:
		return message[T]{}, false

	case msg := <-r.c:
		r.c <- msg
		return msg, true
	}
}

func (r *Receiver[T]) read() message[T] {
	msg := <-r.c
	r.c <- msg

	return msg
}

func (r *Receiver[T]) swap(m message[T]) {
	r.c = m.c
}

func always[T any](cf Cancellable[T]) T {
	v, _ := cf(make(chan struct{}))
	return v
}

// Read a single value that has been broadcast. Blocks until messages have been
// written.
func (r *Receiver[T]) Read() T {
	return always(r.ReadCancel)
}

// Peek reads a single value that has been broadcast and then places it back
// onto the receiver. Blocks until messages have been written.
func (r *Receiver[T]) Peek() T {
	return always(r.PeekCancel)
}

// ReadChan locks the receiver and writes any broadcasted messages to the output
// channel. When you are ready to stop receiving messages, close the second
// signaling channel.
func (r *Receiver[T]) ReadChan() (<-chan T, chan<- struct{}) {
	vc := make(chan T)
	cc := make(chan struct{})

	go func() {
		defer close(vc)

		for {
			result, ok := r.ReadCancel(cc)
			if !ok {
				return
			}

			vc <- result
		}
	}()

	return vc, cc
}

type Cancellable[T any] func(<-chan struct{}) (T, bool)

func (r *Receiver[T]) ReadCancel(c <-chan struct{}) (T, bool) {
	r.mx.Lock()
	defer r.mx.Unlock()

	msg, ok := r.readCancel(c)
	if ok {
		r.swap(msg)
	}

	return msg.v, ok
}

func (r *Receiver[T]) PeekCancel(c <-chan struct{}) (T, bool) {
	r.mx.Lock()
	defer r.mx.Unlock()

	msg, ok := r.readCancel(c)
	return msg.v, ok
}
