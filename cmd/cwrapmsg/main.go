package main

import (
	"fmt"
	"log"

	"github.com/mrsubudei/cwrapmsg"
	"github.com/spf13/pflag"
)

var (
	helpInfo = `Usage: cwrap [OPTION]...

-u     --unstaged   handle only unstaged files
-s     --skip       do not take into account list in cwrap.txt
-h     --help       show help information
-v     --version    print cwrap version
`

	versionInfo = "cwrap version: v1.0.6"
)

func main() {
	var version bool
	pflag.BoolVarP(&version, "version", "v", false, "print cwrap version")

	var help bool
	pflag.BoolVarP(&help, "help", "h", false, "show help information")

	var skip bool
	pflag.BoolVarP(&skip, "skip", "s", false, "do not take into account list in cwrap.txt")

	var unstaged bool
	pflag.BoolVarP(&unstaged, "unstaged", "u", false, "handle only unstaged files")

	pflag.Parse()

	if help {
		fmt.Println(helpInfo)

		return
	}

	if version {
		fmt.Println(versionInfo)

		return
	}

	flags := cwrapmsg.Flags{
		SkipIgnoring: skip,
		OnlyUnstaged: unstaged,
	}

	if err := cwrapmsg.Handle(flags); err != nil {
		log.Fatal(err)
	}
}
