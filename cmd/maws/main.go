package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/fujiwara/maws-cli"
)

func main() {
	opt := maws.Option{}
	flag.StringVar(&opt.Config, "config", "maws.yaml", "path of a config file")
	flag.BoolVar(&opt.BufferStdout, "buffering", true, "buffering stdout of aws cli")
	flag.Int64Var(&opt.MaxParallels, "max-parallels", 10, "max parallels")
	flag.Parse()
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
