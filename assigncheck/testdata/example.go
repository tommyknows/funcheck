package pkg

import "fmt"

func Fn2() {
	{ // simplest case
		x := 5
		fmt.Println(x)
		x = 6 // want `^re-assignment of x$`
		fmt.Println(x)
	}
	{ // more "normal" re-assignments
		x, y := 5, 6
		x, y = 6, x  // want `^re-assignment of x$` `^re-assignment of y$`
		x, z := 5, 6 // want `^re-assignment of x$`
		a, x := 5, 6 // want `^re-assignment of x$`
		fmt.Println(a, x, y, z)

		x++     // want `inline re-assignment of x`
		x += 5  // want `^re-assignment of x$`
		x -= 5  // want `^re-assignment of x$`
		x *= 3  // want `^re-assignment of x$`
		x /= 3  // want `^re-assignment of x$`
		x %= 3  // want `^re-assignment of x$`
		x &= 3  // want `^re-assignment of x$`
		x |= 3  // want `^re-assignment of x$`
		x ^= 3  // want `^re-assignment of x$`
		x <<= 3 // want `^re-assignment of x$`
		x >>= 3 // want `^re-assignment of x$`
		x &^= 3 // want `^re-assignment of x$`
		fmt.Println(x)
	}
	{ // shadowing
		var x int
		{
			x = 2  // want `^re-assignment of x$`
			x := 5 // separate block, re-declaration introduces shadowing
			fmt.Println(x)
		}
		fmt.Println(x)
	}
	{ // type assertions
		type X interface{}
		var z X = 5
		fmt.Println(z)
		if _, ok := z.(int); !ok {
			fmt.Println("error")
		}
		fmt.Println(z)
	}
	{ // assignment of returned function values
		s := test()
		s = test() // want `^re-assignment of s$`
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
	{ // loops, inline re-assignment
		for x := 0; x < 5; x++ { // want `^inline re-assignment of x$`
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
		h = func(i int) int { // want `^re-assignment of h$`
			return i + x
		}

		fmt.Println(h(x))

		var f2 func(int) int
		// this should be valid as there are only comments in between
		f2 = func(i int) int {
			return i + x
		}

		fmt.Println(f2(x))
	}
	{ // edge case function literals
		var g func()
		var f func()
		g = func() { // want `^re-assignment of g$`
			f()
		}

		f = func() {} // want `^re-assignment of f$`

		g()
	}
	{ // block variable definition
		var (
			x int
			f func() int
		)
		f = func() int { return x }
		f()
	}
}

func test() string {
	return "hello"
}
