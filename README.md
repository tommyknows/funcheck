# Funcheck

[![Build Status](https://xn--s68h.ramonr.ch/api/badges/ramon/funcheck/status.svg)](https://xn--s68h.ramonr.ch/ramon/funcheck)

Funcheck is a Go linter that reports non-functional constructs. There is actually
only one checker (`assigncheck`) which reports every case where we only
assign variables (`x = "whatever"`), not combined by a declaration (`x := "whatever"`).
See [my Bachelor thesis](https://github.com/tommyknows/bachelor) for more info
and background.

Currently, there is one exception to the above rule:

- Anonymous function assignments. As it is not possible (because of scoping rules)
  to recursively call an anonymous function without declaring it first, this needed
  to be an exception. The following code block is not valid Go:

```go
func x() {
    y := func(i int) int {
        if i == 0 {
            return -1
        }
        return y(i-1)
    }
}
```

  As the function `y` is not in scope within `y`'s body. To work around this, the
  following code works:

```go
func x() {
    var y func(int) int
    y = func(i int) int {
        if i == 0 {
            return -1
        }
        return y(i-1)
    }
}
```

  This is why this exception exists. However, this exception needs the function's
  declaration to be right before the function's assignment. This is not valid:

```go
func x() {
    var y func(int) int
    z := 5
    y = func(i int) int {
        if i == 0 {
            return -1
        }
        return y(i-1)
    }
}
```
