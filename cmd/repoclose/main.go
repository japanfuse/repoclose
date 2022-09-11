package main

import (
	"repoclose-checker"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(repoclose.Analyzer) }
