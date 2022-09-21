package main

import (
	"github.com/japanfuse/repoclose"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(repoclose.Analyzer) }
