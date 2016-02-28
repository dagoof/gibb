package gibb

import "fmt"

func ExampleReceiver_Read() {
	bc := New()
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
	bc := New()
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
	bc := New()
	r := bc.Listen()

	bc.Write("hello")
	bc.Write(1)
	bc.Write("world")

	var s string
	var n int

	fmt.Println(r.ReadVal(&s), s)
	fmt.Println(r.ReadVal(&s), s)
	fmt.Println(r.ReadVal(&n), n)
	fmt.Println(r.ReadVal(&n), n)
	fmt.Println(r.ReadVal(&s), s)

	// Output:
	// true hello
	// false hello
	// true 1
	// false 1
	// true world
}

func ExampleReceiver_ReadVal_usage() {
	bc := New()
	r := bc.Listen()

	const N = 4
	for i := 0; i < N; i++ {
		bc.Write(i)
	}
	bc.Write("done")

	var n int
	var s string

	for r.ReadVal(&n) {
		fmt.Println(n)
	}
	r.ReadVal(&s)
	fmt.Println(s)

	// Output:
	// 0
	// 1
	// 2
	// 3
	// done
}

func ExampleReceiver_ReadChan() {
	const numWritten = 4

	bc := New()
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
}

func ExampleBroadcaster() {
	bc := New()

	ra := bc.Listen()
	rb := bc.Listen()
	rc := bc.Listen()

	bc.Write(1)
	bc.Write(2)

	fmt.Println("ra:", ra.Read())
	fmt.Println("rb:", rb.Read())
	fmt.Println("rb:", rb.Read())
	fmt.Println("ra:", ra.Read())
	fmt.Println("rc:", rc.Read())
	fmt.Println("rc:", rc.Read())

	// Output:
	// ra: 1
	// rb: 1
	// rb: 2
	// ra: 2
	// rc: 1
	// rc: 2
}

func ExampleReceiver_ReadCancel() {
	const numToWrite = 5
	const numToRead = 3
	var numRead int

	bc := New()
	r := bc.Listen()

	for i := 0; i < numToWrite; i++ {
		bc.Write(i)
	}

	stop := make(chan struct{})
	process := func(v interface{}) {
		numRead++

		if numRead >= numToRead {
			close(stop)
		}

		fmt.Println(v)
	}

	for {
		if result := r.ReadCancel(stop); result.Okay {
			process(result.Val)
			continue
		}

		break
	}

	// Output:
	// 0
	// 1
	// 2
}

func ExampleReceiver_MustReadValCancel() {
	bc := New()
	r := bc.Listen()

	bc.Write("hello")
	bc.Write(1)
	bc.Write("cruel")
	bc.Write(2)
	bc.Write("world")
	bc.Write(true)
	bc.Write("jk")

	var s string
	var count int
	stop := make(chan struct{})

	for r.MustReadValCancel(&s, stop) {
		count++
		fmt.Println(s)

		if count > 2 {
			close(stop)
		}
	}

	// Output:
	// hello
	// cruel
	// world
}
