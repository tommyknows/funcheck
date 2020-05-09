package pkg

import "fmt"

func Fn2() {
	{
		x := 5
		fmt.Println(x)
		x = 6 // want `^re-assignment of x$`
		fmt.Println(x)
	}
	{
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
	{
		var x int
		{
			x = 2  // want `^re-assignment of x$`
			x := 5 // separate block, re-declaration introduces shadowing
			fmt.Println(x)
		}
		fmt.Println(x)
	}
	{
		type X interface{}
		var z X = 5
		fmt.Println(z)
		if _, ok := z.(int); !ok {
			fmt.Println("error")
		}
		fmt.Println(z)
	}
	{
		s := test()
		s = test() // want `^re-assignment of s$`
		fmt.Println(s)
	}
	{
		var f func()
		x, f := 3, func() {
			f()
		}
		fmt.Println(x)
	}
	{
		x, f := 3, func() {}
		fmt.Println(x, f)
	}
	{
		var f func()
		x, f, y := 3, func() {}, 4
		fmt.Println(x, y, f)
	}
}

func Loop() {
	for x := 0; x < 5; x++ { // want `^inline re-assignment of x$`
		fmt.Println(x)
	}

	_ = "hello"
	_ = "world" // blank ident should be ignored
}

func test() string {
	return "hello"
}

func AnonFuncDecl() {
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
