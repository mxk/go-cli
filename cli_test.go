package cli

import (
	"errors"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCmd struct{ run func(args []string) error }

func newTestCmd(run func(args []string) error) func() Cmd {
	return func() Cmd { return &testCmd{run} }
}

func (cmd *testCmd) Main(args []string) error {
	if cmd.run == nil {
		return nil
	}
	return cmd.run(args)
}

func TestCfgAdd(t *testing.T) {
	var main Cfg
	c1 := main.Add(&Cfg{Name: "c1"})
	assert.PanicsWithValue(t, "cli: command already has a parent: c1", func() { (&Cfg{}).Add(c1) })
	assert.PanicsWithValue(t, "cli: missing command name", func() { main.Add(&Cfg{}) })
	assert.PanicsWithValue(t, "cli: missing command name", func() { main.Add(&Cfg{Name: "c2|"}) })
	assert.PanicsWithValue(t, "cli: duplicate command name: c1", func() { main.Add(&Cfg{Name: "x|c1|y"}) })
}

func TestCmd(t *testing.T) {
	var main Cfg
	main.Add(&Cfg{Name: "c1", New: newTestCmd(nil)})
	grp := main.Add(&Cfg{Name: "group|g"})
	grp.Add(&Cfg{
		Name:    "c2",
		MinArgs: 2,
		MaxArgs: 2,
		New:     newTestCmd(nil),
	})
	var mainArgs []string
	grp.Add(&Cfg{
		Name:    "c3",
		MinArgs: 1,
		MaxArgs: 2,
		New: newTestCmd(func(args []string) error {
			mainArgs = append(mainArgs, args...)
			return nil
		}),
	})

	cfg, cmd, args, err := main.Parse(split(""))
	require.NoError(t, err)
	require.IsType(t, &nilCmd{}, cmd)
	require.Empty(t, Name(cfg))
	require.Empty(t, args)

	cfg, cmd, args, err = main.Parse(split("c1"))
	require.NoError(t, err)
	require.IsType(t, &testCmd{}, cmd)
	require.Equal(t, "c1", Name(cfg))
	require.Empty(t, args)

	cfg, cmd, args, err = main.Parse(split("g"))
	require.NoError(t, err)
	require.IsType(t, &nilCmd{}, cmd)
	require.Equal(t, "group", Name(cfg))
	require.Empty(t, args)

	cfg, cmd, args, err = main.Parse(split("group c2 a b"))
	require.NoError(t, err)
	require.IsType(t, &testCmd{}, cmd)
	require.Equal(t, "c2", Name(cfg))
	require.Equal(t, split("a b"), args)

	rc := resetExit()
	oldArgs := os.Args
	os.Args = split("prog g c3 x")
	main.Run()
	os.Args = oldArgs
	assert.Equal(t, 0, *rc)
	assert.Equal(t, split("x"), mainArgs)
	Exit = os.Exit

	_, _, _, err = main.Parse(split("x"))
	assert.EqualError(t, err, `unknown command "x"`)

	_, _, _, err = main.Parse(split("-h"))
	assert.Equal(t, err, flag.ErrHelp)

	_, _, _, err = main.Parse(split("c1 -h"))
	assert.Equal(t, err, flag.ErrHelp)

	_, _, _, err = main.Parse(split("c1 -opt"))
	assert.EqualError(t, err, "flag provided but not defined: -opt")

	_, _, _, err = main.Parse(split("c1 a"))
	assert.EqualError(t, err, "command does not accept any arguments")

	_, _, _, err = main.Parse(split("g c2"))
	assert.EqualError(t, err, "command requires 2 argument(s)")

	_, _, _, err = main.Parse(split("g c3"))
	assert.EqualError(t, err, "command requires at least 1 argument(s)")

	_, _, _, err = main.Parse(split("g c3 a b c"))
	assert.EqualError(t, err, "command accepts at most 2 argument(s)")
}

func TestExit(t *testing.T) {
	defer func() { Exit = os.Exit }()
	var err error
	cfg := Cfg{New: newTestCmd(func([]string) error { return err })}
	noArgs := make([]string, 0)

	Bin = "bin"
	err = ErrHelp
	done, rc := interceptWrite(&os.Stderr), resetExit()
	cfg.Run(noArgs...)
	assert.Equal(t, "Usage: bin\n       bin help\n\n", done())
	assert.Equal(t, 2, *rc)

	err = UsageError("usage error")
	done, rc = interceptWrite(&os.Stderr), resetExit()
	cfg.Run(noArgs...)
	assert.Equal(t, "Error: usage error\nUsage: bin\n       bin help\n", done())
	assert.Equal(t, 2, *rc)

	err = ExitCode(42)
	done, rc = interceptWrite(&os.Stderr), resetExit()
	cfg.Run(noArgs...)
	assert.Equal(t, "", done())
	assert.Equal(t, 42, *rc)

	err = errors.New("fail")
	done, rc = interceptWrite(&os.Stderr), resetExit()
	cfg.Run(noArgs...)
	assert.Equal(t, "Error: fail\n", done())
	assert.Equal(t, 1, *rc)
}

func TestNilCmd(t *testing.T) {
	defer func() { Exit = os.Exit }()
	var main Cfg

	done, rc := interceptWrite(&os.Stderr), resetExit()
	New(&main).Main(nil)
	assert.Equal(t, "Command not implemented\n", done())
	assert.Equal(t, 2, *rc)

	main.Add(&Cfg{Name: "c1"})
	main.Add(&Cfg{Name: "c2", Summary: "Command 2"})
	done, rc = interceptWrite(&os.Stderr), resetExit()
	New(&main).Main(nil)
	assert.Equal(t, Dedent(`
		Specify command:
		  c1
		  c2  Command 2
	`)[1:], done())
	assert.Equal(t, 2, *rc)
}

func split(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, " ")
}

func resetExit() *int {
	rc := -1
	Exit = func(code int) {
		Exit = func(code int) {}
		rc = code
	}
	return &rc
}
