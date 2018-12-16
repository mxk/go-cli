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

type testCmd struct {
	ci  *Info
	run func(args []string) error
}

func newTestCmd(ci *Info, run func(args []string) error) func() Cmd {
	return func() Cmd { return &testCmd{ci, run} }
}

func (cmd *testCmd) Info() *Info { return cmd.ci }

func (cmd *testCmd) Main(args []string) error {
	if cmd.run == nil {
		return nil
	}
	return cmd.run(args)
}

func TestCmd(t *testing.T) {
	var c1, c2, c3 Info
	c1 = *Main.Add(&Info{Name: "c1", New: newTestCmd(&c1, nil)})
	grp := Main.Add(&Info{Name: "group|g"})
	c2 = *grp.Add(&Info{
		Name:    "c2",
		MinArgs: 2,
		MaxArgs: 2,
		New:     newTestCmd(&c2, nil),
	})
	var mainArgs []string
	c3 = *grp.Add(&Info{
		Name:    "c3",
		MinArgs: 1,
		MaxArgs: 2,
		New: newTestCmd(&c3, func(args []string) error {
			mainArgs = append(mainArgs, args...)
			return nil
		}),
	})

	cmd, args, err := Main.Parse(split(""))
	require.NoError(t, err)
	require.IsType(t, &nilCmd{}, cmd)
	require.Empty(t, cmd.Info().PrimaryName())
	require.Empty(t, args)

	cmd, args, err = Main.Parse(split("c1"))
	require.NoError(t, err)
	require.IsType(t, &testCmd{}, cmd)
	require.Equal(t, "c1", cmd.Info().PrimaryName())
	require.Empty(t, args)

	cmd, args, err = Main.Parse(split("g"))
	require.NoError(t, err)
	require.IsType(t, &nilCmd{}, cmd)
	require.Equal(t, "group", cmd.Info().PrimaryName())
	require.Empty(t, args)

	cmd, args, err = Main.Parse(split("group c2 a b"))
	require.NoError(t, err)
	require.IsType(t, &testCmd{}, cmd)
	require.Equal(t, "c2", cmd.Info().PrimaryName())
	require.Equal(t, split("a b"), args)

	rc := resetExit()
	Main.Run(split("g c3 x"))
	assert.Equal(t, 0, *rc)
	assert.Equal(t, split("x"), mainArgs)
	Exit = os.Exit

	_, _, err = Main.Parse(split("x"))
	assert.EqualError(t, err, `unknown command "x"`)

	_, _, err = Main.Parse(split("-h"))
	assert.EqualError(t, err, flag.ErrHelp.Error())

	_, _, err = Main.Parse(split("c1 -h"))
	assert.EqualError(t, err, flag.ErrHelp.Error())

	_, _, err = Main.Parse(split("c1 -opt"))
	assert.EqualError(t, err, "flag provided but not defined: -opt")

	_, _, err = Main.Parse(split("c1 a"))
	assert.EqualError(t, err, "command does not accept any arguments")

	_, _, err = Main.Parse(split("g c2"))
	assert.EqualError(t, err, "command requires 2 argument(s)")

	_, _, err = Main.Parse(split("g c3"))
	assert.EqualError(t, err, "command requires at least 1 argument(s)")

	_, _, err = Main.Parse(split("g c3 a b c"))
	assert.EqualError(t, err, "command accepts at most 2 argument(s)")
}

func TestExit(t *testing.T) {
	defer func() { Exit = os.Exit }()
	var ci Info
	var err error
	ci = Info{New: newTestCmd(&ci, func([]string) error { return err })}

	Bin = "bin"
	err = flag.ErrHelp
	done, rc := interceptWrite(&os.Stderr), resetExit()
	ci.Run(nil)
	assert.Equal(t, "Usage: bin\n       bin help\n\n", done())
	assert.Equal(t, 2, *rc)

	err = UsageError("usage error")
	done, rc = interceptWrite(&os.Stderr), resetExit()
	ci.Run(nil)
	assert.Equal(t, "Error: usage error\nUsage: bin\n       bin help\n", done())
	assert.Equal(t, 2, *rc)

	err = ExitCode(42)
	done, rc = interceptWrite(&os.Stderr), resetExit()
	ci.Run(nil)
	assert.Equal(t, "", done())
	assert.Equal(t, 42, *rc)

	err = errors.New("fail")
	done, rc = interceptWrite(&os.Stderr), resetExit()
	ci.Run(nil)
	assert.Equal(t, "Error: fail\n", done())
	assert.Equal(t, 1, *rc)
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
