package main

import (
	"fmt"
	"log"

	"github.com/mrsubudei/cwrapmsg"
	"github.com/spf13/pflag"
)

var (
	versionInfo = "cwrapmsg version: v0.0.1"
)

func main() {
	var version bool
	pflag.BoolVarP(&version, "version", "v", false, "print cwrapmsg version")

	pflag.Parse()

	if version {
		fmt.Println(versionInfo)

		return
	}

	if err := cwrapmsg.Handle(); err != nil {
		log.Fatal(err)
	}
}
