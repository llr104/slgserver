// +build !windows,!plan9

package config

import (
	"os"
	"os/signal"
	"syscall"
)

func signalReload() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)

	for {
		sig := <-ch
		switch sig {
		case syscall.SIGUSR1:
			ReloadConfigFile()
		}
	}
}
