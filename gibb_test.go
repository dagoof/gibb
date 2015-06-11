package gibb

import (
	"fmt"
	"sync"
	"time"
)

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

func ExampleReceiver_ReadVal_multiplex() {
	type done struct{}

	bc := New()
	rs := bc.Listen()
	ri := bc.Listen()

	bc.Write(1)
	bc.Write(2)
	bc.Write("a")
	bc.Write("b")
	bc.Write(3)
	bc.Write("c")

	bc.Write(done{})

	var n int
	var s string
	var d done

	for !rs.ReadVal(&d) {
		if !rs.ReadVal(&s) {
			rs.Read()
			continue
		}

		fmt.Println("string stream", s)
	}

	for !ri.ReadVal(&d) {
		if !ri.ReadVal(&n) {
			ri.Read()
			continue
		}

		fmt.Println("int stream", n)
	}

	// Output:
	// string stream a
	// string stream b
	// string stream c
	// int stream 1
	// int stream 2
	// int stream 3
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

func ExampleReadTimeout() {
	bc := New()

	r := bc.Listen()

	bc.Write(1)
	bc.Write(2)

	fmt.Println(r.Read())
	fmt.Println(r.ReadTimeout(time.Second).Describe())

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		fmt.Println(r.ReadTimeout(time.Millisecond * 10).Describe())
		wg.Done()
	}()

	go func() {
		fmt.Println(r.ReadTimeout(time.Millisecond * 50).Describe())
		wg.Done()
	}()

	time.Sleep(time.Millisecond * 15)
	bc.Write(3)
	wg.Wait()

	// Output:
	// 1
	// just 2
	// nothing
	// just 3
}

func ExampleReadValCancel() {
	bc := New()
	r := bc.Listen()

	bc.Write(1)
	bc.Write("hello")
	bc.Write(2)

	var s string
	var n int

	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &s).Describe())
	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &n).Describe())
	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &n).Describe())
	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &s).Describe())
	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &s).Describe())
	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &n).Describe())
	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &n).Describe())
	fmt.Println(r.ReadValCancel(Timeout(time.Millisecond), &s).Describe())

	// Output:
	// invalid
	// just 1
	// invalid
	// just hello
	// invalid
	// just 2
	// invalid
	// invalid
}
