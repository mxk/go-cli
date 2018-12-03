package cli

import (
	"os"
	"os/signal"
	"sync"
)

var signalsOnce sync.Once

// ExitSignals returns signals for which the default behavior is to exit without
// a stack dump. It must be called before signal handlers are modified.
func ExitSignals() []os.Signal {
	signalsOnce.Do(func() {
		keep := exitSignals[:0]
		for _, s := range exitSignals {
			if !signal.Ignored(s) {
				keep = append(keep, s)
			}
		}
		if len(keep) == 0 {
			// signal functions operate on all signals if none are provided, so
			// we should always return at least one exit signal.
			keep = append(keep, os.Interrupt)
		}
		exitSignals = keep
	})
	return exitSignals
}
