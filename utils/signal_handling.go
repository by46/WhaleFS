package utils

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	signalChan chan os.Signal
	hooks      = make([]func(), 0)
	hookLock   sync.Mutex
)

func init() {
	signalChan = make(chan os.Signal, 0)
	signal.Ignore(syscall.SIGHUP)
	signal.Notify(signalChan,
		os.Interrupt,
		os.Kill,
		syscall.SIGALRM,
		syscall.SIGTERM,
		syscall.SIGINT)

	go func() {
		for _ = range signalChan {
			for _, hook := range hooks {
				hook()
			}
			os.Exit(0)
		}
	}()
}

func OnInterrupt(fn func()) {
	hookLock.Lock()
	defer hookLock.Unlock()

	// deal with control+c,etc
	// controlling terminal close, daemon not exit
	hooks = append(hooks, fn)
}
