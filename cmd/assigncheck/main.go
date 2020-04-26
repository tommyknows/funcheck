package main

import (
	"git.ramonr.ch/ramon/bachelor/work/funcheck/assigncheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(assigncheck.Analyzer)
}
