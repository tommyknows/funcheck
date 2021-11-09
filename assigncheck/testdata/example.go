package pkg

import (
	"encoding/base64"
	"fmt"
)

func Test() {
	{ // simplest case
		x := 5
		fmt.Println(x)
		x = 6 // want `^reassignment of x$`
		fmt.Println(x)
	}
	{ // more "normal" reassignments
		x, y := 5, 6
		x, y = 6, x  // want `^reassignment of x$` `^reassignment of y$`
		x, z := 5, 6 // want `^reassignment of x$`
		a, x := 5, 6 // want `^reassignment of x$`
		fmt.Println(a, x, y, z)

		x++     // want `inline reassignment of x`
		x += 5  // want `^reassignment of x$`
		x -= 5  // want `^reassignment of x$`
		x *= 3  // want `^reassignment of x$`
		x /= 3  // want `^reassignment of x$`
		x %= 3  // want `^reassignment of x$`
		x &= 3  // want `^reassignment of x$`
		x |= 3  // want `^reassignment of x$`
		x ^= 3  // want `^reassignment of x$`
		x <<= 3 // want `^reassignment of x$`
		x >>= 3 // want `^reassignment of x$`
		x &^= 3 // want `^reassignment of x$`
		fmt.Println(x)
	}
	{ // shadowing
		var x int
		{
			x = 2  // want `^reassignment of x$`
			x := 5 // separate block, re-declaration introduces shadowing
			fmt.Println(x)
		}
		fmt.Println(x)
	}
	{ // inline assignment
		if r := recover(); r != nil {
			fmt.Println("recoverd", r)
		}
	}
	{ // type assertions
		type X interface{}
		var z X = 5
		fmt.Println(z)
		if _, ok := z.(int); !ok {
			fmt.Println("error")
		}
		var ok bool
		_, ok = z.(int) // want `^reassignment of ok$`
		if !ok {
			fmt.Println("error")
		}
		fmt.Println(z)
	}
	{ // assignment of returned function values
		s := test()
		s = test() // want `^reassignment of s$`
		fmt.Println(s)
	}
	{ // function literals paired with redeclarations
		var f func()
		x, f := 3, func() {
			f()
		}
		fmt.Println(x)
	}
	{ // declaration of a function literal
		x, f := 3, func() {}
		fmt.Println(x, f)
	}
	{ // more edge case function literals
		var f func()
		x, f, y := 3, func() {}, 4
		fmt.Println(x, y, f)
	}
	{ // loops, inline reassignment
		for x := 0; x < 5; x++ { // want `reassignment \(for loop\) in "for x := 0; x < 5; x\+\+ { ... }` `inline reassignment`
			fmt.Println(x)
		}
	}
	{ // blank identifier
		_ = "hello"
		_ = "world" // blank ident should be ignored
	}
	{ // more function literals
		var f func(i int) int
		f = func(i int) int {
			if i == 0 {
				return -1
			}
			return f(i - 1)
		}

		var g func(int) int = func(x int) int {
			return f(x)
		}

		fmt.Println(g(3))

		var h func(int) int
		x := 5
		// this should be invalid as there are statements in between
		h = func(i int) int { // want `^reassignment of h$`
			return i + x
		}

		fmt.Println(h(x))

		var f2 func(int) int
		// this should be valid as there are only comments in between
		f2 = func(i int) int {
			return i + x
		}

		fmt.Println(f2(x))

		var g1, g2 func()
		g1, g2 = func() { g2() }, func() { g1() }

		var h1 func()
		var h2 func()
		h1, h2 = func() { h2() }, func() { h1() } // want `^reassignment of h1$`
	}
	{ // edge case function literals
		var g func()
		var f func()
		g = func() { // want `^reassignment of g$`
			f()
		}

		f = func() {} // want `^reassignment of f$`

		g()

		var h1 func()
		h1, h2 := func() { h1() }, func() { h1() }
		h2()
	}
	{ // block variable definition
		var (
			x int
			f func() int
		)
		f = func() int { return x }
		f()
	}
	{ // variable declared in different file
		declaredInDifferentFile = 5 // want `^reassignment of declaredInDifferentFile$`
		fmt.Println(declaredInDifferentFile)
	}
	{ // variable declared in different package
		base64.StdEncoding = base64.NewEncoding("") // want `^reassignment of base64.StdEncoding$`
	}
	{ // map and slice access
		m := map[string]string{
			"hello": "world",
		}
		m["test"] = "no" // want `^reassignment of m\["test"\]$`
		s := []int{1, 2, 3}
		s[0] = 2 // want `^reassignment of s\[0\]$`
	}
	{ // simple for loop
		x := []int{1, 2, 3}
		for i := range x { // want `^internal reassignment \(for loop\) in "for i := range x { ... }"$`
			fmt.Println(i)
		}
	}
	{ // allow forever loops
		for {
			fmt.Println("hello")
		}
	}
	{ // loop without a modification
		for i := 3; i != 2; {
			fmt.Println("hi")
		}
	}
}

func test() string {
	return "hello"
}
