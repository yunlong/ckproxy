package main

import (
	"fmt"
	"os"
)

var (
	AUTO_BUILD_VERSION     = "0.0.0.0"
	AUTO_BUILD_TIME        = "1970-01-01 00:00:00"
	AUTO_BUILD_COMMIT_SHA1 = "unknown"
)

func init() {
	arg_num := len(os.Args)
	if arg_num > 1 {
		for i := 1; i < arg_num; i++ {
			if os.Args[i] == `-version` {
				fmt.Println("Version: ", AUTO_BUILD_VERSION)
				fmt.Println("Git Tag: ", AUTO_BUILD_COMMIT_SHA1)
				fmt.Println("Build Time: ", AUTO_BUILD_TIME)
				os.Exit(0)
			}
		}
	}
}
