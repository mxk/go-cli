package cli

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitSignals(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, []os.Signal{os.Interrupt}, ExitSignals())
		defer signal.Reset(os.Interrupt)

		signal.Ignore(os.Interrupt)
		signalsOnce = sync.Once{}
		assert.Equal(t, []os.Signal{os.Interrupt}, ExitSignals())
		return
	}

	want := append([]os.Signal(nil), exitSignals...)
	assert.Equal(t, want, ExitSignals())
	defer signal.Reset(want...)

	want, last := want[:len(want)-1], want[len(want)-1]
	signal.Ignore(last)
	signalsOnce = sync.Once{}
	assert.Equal(t, want, ExitSignals())

	signal.Ignore(want...)
	signalsOnce = sync.Once{}
	assert.Equal(t, []os.Signal{os.Interrupt}, ExitSignals())
}
