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

func ExampleReceiver_ReadVal() {
	bc := NewBroadcaster()
	r := bc.Listen()

	bc.Write("hello")
	bc.Write(1)
	bc.Write("world")

	var s string

	r.ReadVal(&s)
	fmt.Println(s)
	r.ReadVal(&s)
	fmt.Println(s)

	// Output:
	// hello
	// world
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
