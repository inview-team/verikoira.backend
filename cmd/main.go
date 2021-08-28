package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inview-team/verikoira.backend/api"
	"github.com/inview-team/verikoira.backend/internal/config"
	"github.com/inview-team/verikoira.backend/internal/logger"
	"go.uber.org/zap"
)

var cfgFile string

func init() {
	flag.StringVar(&cfgFile, "config", "configs/example_config.json", "path to config file")
}

func main() {
	flag.Parse()
	conf, err := config.Init(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	err = logger.InitLogger(conf.Logger.File, conf.Logger.Level)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println(conf.Rmq.Address)
	log.Println(conf.Rmq.ReadQueue)

	k, err := api.New(conf, ctx)
	if err != nil {
		zap.L().Error("cannot initialize API", zap.Error(err))
		return
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go k.Run()

	<-done
	signal.Stop(done)
}
