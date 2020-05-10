# PrettyPrint

This directory contains the `PrettyPrint' package. It visualises a
Go AST, but only Assignment Nodes. In these nodes, it prints their
position, the identifiers and the identifier's declaration plus position.

It is used to visualise how assigncheck is checking for re-assignments.

To see it for yourself, run `go test -v .` in this directory.

```
=== RUN   TestRun
Assignment "x, y := 1, 2": 2958101
        Ident "x": 2958101
                Decl "x, y := 1, 2": 2958101
        Ident "y": 2958104
                Decl "x, y := 1, 2": 2958101
Assignment "y = 3": 2958115
        Ident "y": 2958115
                Decl "x, y := 1, 2": 2958101
--- PASS: TestRun (0.55s)
PASS
ok      github.com/tommyknows/funcheck/prettyprint      (cached)
```
