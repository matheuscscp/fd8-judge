package helpers

import (
	"os"
	osSignal "os/signal"
	"syscall"
)

// HandleInterrupt blocks until a signal is caught, when it calls f(signal); if f(signal) returns
// false, HandleInterrupt remains blocking until another signal is caught, when the same condition
// is checked again.
func HandleInterrupt(f func(os.Signal) bool) {
	signalChannel := make(chan os.Signal, 1)
	osSignal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	for {
		signal := <-signalChannel
		if f(signal) {
			osSignal.Stop(signalChannel)
			return
		}
	}
}
