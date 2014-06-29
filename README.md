gibb
====

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

[![Build Status](https://drone.io/github.com/dagoof/gibb/status.png)](https://drone.io/github.com/dagoof/gibb/latest)

https://godoc.org/github.com/dagoof/gibb
