package main

import (
	"github.com/tommyknows/funcheck/assigncheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(assigncheck.Analyzer)
}
