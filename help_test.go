package cli

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type helpCmd struct{ Opt string `cli:"Option description"` }

func (*helpCmd) Main(args []string) error { return nil }
func (*helpCmd) Help(w *Writer) {
	w.Text("Command help.")
	w.Text("Next paragraph.\n\n\n")
}

func TestHelp(t *testing.T) {
	g := Cfg{
		Summary: "Group",
		New:     newTestCmd(nil),
	}
	c1 := g.Add(&Cfg{
		Name:    "c1",
		Summary: "Command 1",
		New:     newTestCmd(nil),
	})
	c2 := g.Add(&Cfg{Name: "c2"})
	c3 := c2.Add(&Cfg{
		Name:    "c3|c",
		Usage:   "usage",
		Summary: "Command 3",
		New:     func() Cmd { return &helpCmd{} },
	})
	Bin = "bin"
	assert.Equal(t, Dedent(`
		Usage: bin <command> [options] ...
		       bin <command> help
		       bin help [command]

		Group.

		Commands:
		  c1  Command 1
		  c2

	`)[1:], g.Help().String())
	assert.Equal(t, Dedent(`
		Usage: bin c1
		       bin c1 help

		Command 1.

	`)[1:], c1.Help().String())
	assert.Equal(t, Dedent(`
		Usage: bin c2 <command> [options] ...
		       bin c2 <command> help
		       bin c2 help [command]

		Commands:
		  c3  Command 3

	`)[1:], c2.Help().String())
	assert.Equal(t, Dedent(`
		Usage: bin c2 {c3|c} usage
		       bin c2 {c3|c} help

		Command help.

		Next paragraph.

		Options:
		  -opt string
		    	Option description

	`)[1:], c3.Help().String())
}

func TestDedent(t *testing.T) {
	tests := []*struct{ in, out string }{
		// Pass-through
		{"", ""},
		{"A", "A"},
		{"\t", "\t"},
		{"\tA", "\tA"},
		{"\n", "\n"},
		{"\nA", "\nA"},
		{"A\nB", "A\nB"},
		{"\t\n", "\t\n"},
		{"\t\nA", "\t\nA"},
		{"\t\n\t", "\t\n\t"},
		{"\n\t\n", "\n\t\n"},
		{"\n\t\n\t", "\n\t\n\t"},
		{"\n\t\t\nA", "\n\t\t\nA"},
		{"\n\t\nA\n\tB", "\n\t\nA\n\tB"},
		{"\t\n\t\t\nA\n\tB\n", "\t\n\t\t\nA\n\tB\n"},

		// Dedent
		{"\n\tA", "\nA"},
		{"\t\n\tA", "\t\nA"},
		{"\n\n\tA", "\n\nA"},
		{"\n\tA\nB", "\nA\nB"},
		{"\nA\n\tB", "\nA\n\tB"},
		{"\n\tA\n\t\tB", "\nA\n\tB"},
		{"\n\t\t\n\tA\n", "\n\t\nA\n"},
		{"A\n\t\tB\nC\n\tD\n\t\t\tE\n\t", "A\nB\nC\nD\n\tE\n"},
	}
	addr := func(s string) uintptr {
		return (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	}
	for _, tc := range tests {
		if s := Dedent(tc.in); assert.Equal(t, tc.out, s) && tc.in == tc.out {
			assert.Equal(t, addr(tc.in), addr(s)) // No allocation
		}
	}
}

func TestIsHelp(t *testing.T) {
	assert.False(t, isHelp(""))
	assert.True(t, isHelp("help"))
}
