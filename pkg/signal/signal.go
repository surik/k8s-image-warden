package signal

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

const DefaultWaitTimeout = 5

func WaitForSignals(shutdownFunc func(), timeout int, sigs ...os.Signal) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, sigs...)
	gotSignal := <-quit

	log.Printf("Shutdown server because of %s signal...\n", gotSignal.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	done := make(chan bool, 1)

	go func() {
		shutdownFunc()
		done <- true
	}()

	select {
	case <-ctx.Done():
		log.Printf("Timeout of %d seconds.", timeout)
	case <-done:
	}

	log.Println("Server was shutdown.")
}
