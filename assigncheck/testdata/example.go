package pkg

import "fmt"

func Fn() {
	x := 5
	fmt.Println(x)
	x = 6 // want `re-assignment of x`
	fmt.Println(x)
}

func Fn2() {
	x, y := 5, 6
	fmt.Println(x, y)
	x, y = 6, x // want `re-assignment of x, y`
	fmt.Println(x, y)

	x++ // want `inline re-assignment of x`
	fmt.Println(x)
	x += 5 // want `re-assignment of x`
	fmt.Println(x)
	x -= 5  // want `re-assignment of x`
	x *= 3  // want `re-assignment of x`
	x /= 3  // want `re-assignment of x`
	x %= 3  // want `re-assignment of x`
	x &= 3  // want `re-assignment of x`
	x |= 3  // want `re-assignment of x`
	x ^= 3  // want `re-assignment of x`
	x <<= 3 // want `re-assignment of x`
	x >>= 3 // want `re-assignment of x`
	x &^= 3 // want `re-assignment of x`

	fmt.Println(x)
	{ // separate block, re-declaration / shadowing
		x := 5
		fmt.Println(x)
	}

	type X interface{}
	var z X = 5
	fmt.Println(z)
	// this is okay as we basically only
	// "ensure" that z has the type int,
	// no conversion or change is made.
	if _, ok := z.(int); !ok {
		fmt.Println("error")
	}
	fmt.Println(z)

	s := test()
	fmt.Println(s)
	s = test() // want `re-assignment of s`
	fmt.Println(s)
}

func Loop() {
	for x := 0; x < 5; x++ { // want `inline re-assignment of x`
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
	h = func(i int) int { // want `re-assignment of h`
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
