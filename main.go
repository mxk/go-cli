// Package cli provides tools for defining command line interfaces.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Bin is the name of the executing binary.
var Bin = filepath.Base(os.Args[0])

// Debug determines whether to print debugging information.
var Debug = false

// Exit is called by Cfg.Run() to terminate the process.
var Exit = os.Exit

// Main is the common root of all commands in a CLI program. It is normally
// called as follows:
//
//	func main() { cli.Main.Run() }
var Main Cfg

// DebugFromEnv sets the Debug flag from the specified environment variable.
func DebugFromEnv(key string) {
	if v, ok := os.LookupEnv(key); v == "" {
		Debug = ok
	} else {
		Debug, _ = strconv.ParseBool(v)
	}
}

// UsageError reports a problem with command options or positional arguments.
type UsageError string

// Error implements error interface.
func (e UsageError) Error() string { return string(e) }

// Error returns a new UsageError, formatting arguments via fmt.Sprintln.
func Error(v ...interface{}) UsageError {
	if len(v) == 1 {
		if s, ok := v[0].(string); ok {
			return UsageError(s)
		}
	}
	s := fmt.Sprintln(v...)
	return UsageError(s[:len(s)-1])
}

// Errorf returns a new UsageError, formatting arguments via fmt.Sprintf.
func Errorf(format string, v ...interface{}) UsageError {
	return UsageError(fmt.Sprintf(format, v...))
}

// ExitCode is an error that sets the exit code without printing any message.
type ExitCode int

// Error implements error interface.
func (e ExitCode) Error() string {
	return fmt.Sprintf("exit code %d", int(e))
}

// Sum returns the number of arguments that are true. This can be used to test
// for mutually exclusive flags.
func Sum(v ...bool) int {
	var n int
	for _, b := range v {
		if b {
			n++
		}
	}
	return n
}
