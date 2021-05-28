package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fujiwara/maws-cli"
	"github.com/hashicorp/logutils"
)

var version string

var filter = &logutils.LevelFilter{
	Levels:   []logutils.LogLevel{"debug", "info", "warn", "error"},
	MinLevel: logutils.LogLevel("info"),
	Writer:   os.Stderr,
}

func main() {
	var showVersion, debug bool
	opt := maws.Option{}
	flag.StringVar(&opt.Config, "config", "maws.yaml", "path of a config file")
	flag.BoolVar(&debug, "debug", false, "enable debug log")
	flag.BoolVar(&opt.BufferStdout, "buffering", true, "buffering stdout of aws cli")
	flag.Int64Var(&opt.MaxParallels, "max-parallels", 10, "max parallels")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Println("maws version", version)
		return
	}
	if debug {
		filter.MinLevel = logutils.LogLevel("debug")
	}
	log.SetOutput(filter)

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
