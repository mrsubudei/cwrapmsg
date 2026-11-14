package main

import (
	"github.com/mrsubudei/cwrapmsg"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(cwrapmsg.NewAnalyzer(nil))
}
