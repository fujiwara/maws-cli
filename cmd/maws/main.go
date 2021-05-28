package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fujiwara/maws-cli"
)

var version string

func main() {
	var showVersion bool
	opt := maws.Option{}
	flag.StringVar(&opt.Config, "config", "maws.yaml", "path of a config file")
	flag.BoolVar(&opt.BufferStdout, "buffering", true, "buffering stdout of aws cli")
	flag.Int64Var(&opt.MaxParallels, "max-parallels", 10, "max parallels")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Println("maws version", version)
		return
	}

	opt.Commands = flag.Args()
	errCount, err := maws.Run(context.Background(), opt)
	if err != nil {
		log.Println("[error]", err)
		os.Exit(1)
	}
	if errCount > 0 {
		log.Printf("[error] %d errors", errCount)
		os.Exit(int(errCount))
	}
}
