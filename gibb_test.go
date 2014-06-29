package gibb

import "fmt"

func ExampleReceiver_Read() {
	bc := NewBroadcaster()
	r := bc.Listen()

	bc.Write("hello")
	bc.Write(1)
	bc.Write("world")

	fmt.Println(r.Read())
	fmt.Println(r.Read())
	fmt.Println(r.Read())

	// Output:
	// hello
	// 1
	// world
}

func ExampleReceiver_MustReadVal() {
	bc := NewBroadcaster()
	r := bc.Listen()

	bc.Write("hello")
	bc.Write(1)
	bc.Write("world")

	var s string

	r.MustReadVal(&s)
	fmt.Println(s)
	r.MustReadVal(&s)
	fmt.Println(s)

	// Output:
	// hello
	// world
}

func ExampleReceiver_ReadVal() {
	bc := NewBroadcaster()
	r := bc.Listen()

	bc.Write("hello")
	bc.Write(1)
	bc.Write("world")

	var s string

	fmt.Println(r.ReadVal(&s), s)
	fmt.Println(r.ReadVal(&s), s)
	fmt.Println(r.ReadVal(&s), s)

	// Output:
	// true hello
	// false hello
	// true world
}

func ExampleReceiver_ReadVal_usage() {
	bc := NewBroadcaster()
	r := bc.Listen()

	const N = 4
	for i := 0; i < N; i++ {
		bc.Write(i)
	}
	bc.Write("done")

	var n int

	for r.ReadVal(&n) {
		fmt.Println(n)
	}

	// Output:
	// 0
	// 1
	// 2
	// 3
}

func ExampleReceiver_ReadChan() {
	const numWritten = 5

	bc := NewBroadcaster()
	r := bc.Listen()

	for i := 0; i < numWritten; i++ {
		bc.Write(i)
	}

	vc, done := r.ReadChan()

	var n int
	for v := range vc {
		n++
		fmt.Println(v)

		if n >= numWritten {
			close(done)
		}
	}

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}
