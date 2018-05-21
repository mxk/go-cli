package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// Bin is the name of the executing binary.
var Bin = filepath.Base(os.Args[0])

// Exit is called by Info.Run() to terminate the process.
var Exit = os.Exit

// Main is the root of all registered commands without an explicit parent. It is
// normally called from main() as follows:
//
//	func main() {
//		cli.Main.Run(os.Args[1:])
//	}
var Main Info

func init() { Main.New = func() Cmd { return (*nilCmd)(&Main) } }

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

// nilCmd implements Cmd interface for commands without their own constructor.
type nilCmd Info

func (cmd *nilCmd) Info() *Info { return (*Info)(cmd) }

func (cmd *nilCmd) Main(args []string) error {
	if cmd.cmds != nil {
		return Error("command not specified")
	}
	return Error("command not implemented")
}
