// +build !windows

package cli

import (
	"os"
	"syscall"
)

var exitSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM}
