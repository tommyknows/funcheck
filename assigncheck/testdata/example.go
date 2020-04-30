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
	x -= 5 // want `re-assignment of x`
	fmt.Println(x)
	{ // separate block, re-declaration / shadowing
		x := 5
		fmt.Println(x)
	}

	type X interface{}
	var z X = 5
	fmt.Println(z)
	var ok bool
	// this is okay as we basically only
	// "ensure" that z has the type int,
	// no conversion or change is made.
	z, ok = z.(int)
	if !ok {
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
}

func test() string {
	return "hello"
}
