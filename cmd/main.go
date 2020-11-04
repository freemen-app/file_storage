package main

import (
	"os"
	"os/signal"

	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/infrastructure/app"
	"github.com/freemen-app/file_storage/infrastructure/events"
	grpcApi "github.com/freemen-app/file_storage/infrastructure/grpc"
)

func main() {
	conf := config.New(config.DefaultConfig)
	application := app.New(conf)
	if err := application.Start(); err != nil {
		panic(err)
	}

	api := grpcApi.New(application, &conf.Api)
	amqp := events.New(application, &conf.AMQP)
	if err := amqp.Start(); err != nil {
		panic(err)
	}
	go api.Start()
	// Wait for interrupt signal to gracefully shutdown the server with
	// api timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	api.Shutdown()
	application.Shutdown()
}
