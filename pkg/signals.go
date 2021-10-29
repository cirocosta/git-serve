package pkg

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func SignalHandlingContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("signal received - cancelling execution")

		cancel()
		<-c
		log.Info("signal received - aborting")
		os.Exit(1) // second signal - force exit.
	}()

	return ctx
}
