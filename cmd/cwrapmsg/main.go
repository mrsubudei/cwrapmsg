package main

import (
	"fmt"
	"log"

	"github.com/mrsubudei/cwrapmsg"
	"github.com/spf13/pflag"
)

var (
	helpInfo = `Usage: cwrapmsg [OPTION]...

-u     --unstaged   handle only unstaged files
-c     --cyrillic   check messages with cyrillic symbols
-s     --skip       do not take into account list in cwrapmsg.txt
-h     --help       show help information
-v     --version    print cwrapmsg version
`

	versionInfo = "cwrapmsg version: v1.0.0"
)

func main() {
	var version bool
	pflag.BoolVarP(&version, "version", "v", false, "print cwrapmsg version")

	var help bool
	pflag.BoolVarP(&help, "help", "h", false, "show help information")

	var skip bool
	pflag.BoolVarP(&skip, "skip", "s", false, "do not take into account list in cwrapmsg.txt")

	var unstaged bool
	pflag.BoolVarP(&unstaged, "unstaged", "u", false, "handle only unstaged files")

	var cyrillic bool
	pflag.BoolVarP(&cyrillic, "cyrillic", "c", false, "check messages with cyrillic symbols")

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
		SkipIgnoring:  skip,
		OnlyUnstaged:  unstaged,
		CheckCyrillic: cyrillic,
	}

	if err := cwrapmsg.Handle(flags); err != nil {
		log.Fatal(err)
	}
}
