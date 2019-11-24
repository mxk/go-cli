package cli

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Cmd is an executable command that receives positional arguments in args.
type Cmd interface{ Main(args []string) error }

// Helper is an optional command interface for providing help information.
type Helper interface {
	Cmd
	Help(w *Writer)
}

// Cfg contains command configuration. Commands are typically defined by
// assigning the value returned by Cfg.Add() to a global variable, which can
// then be used to define sub-commands:
//
//	var exampleCLI = cli.Main.Add(&cli.Cfg{
//		Name: "example|ex",
//		New:  func() cli.Cmd { return &exampleCmd{Opt: "default-value"} },
//	})
//
//	type exampleCmd struct{ Opt string `cli:"Option description"` }
//
//	func (cmd *exampleCmd) Main(args []string) error { return nil }
type Cfg struct {
	Name    string     // '|'-separated command name and optional aliases
	Usage   string     // Option and argument syntax
	Summary string     // Capitalized one-line description without trailing period
	MinArgs int        // Minimum number of positional arguments
	MaxArgs int        // Maximum number of positional arguments
	Hide    bool       // Hide from command list
	New     func() Cmd // Constructor (optional for parent commands)

	parent *Cfg            // Parent command
	cmds   map[string]*Cfg // Sub-commands
}

// New returns a new command for config c.
func New(c *Cfg) Cmd {
	if c.New != nil {
		return c.New()
	}
	return (*nilCmd)(c)
}

// nameSep is the command name separator.
const nameSep = '|'

// Name returns the first entry in c.Name.
func Name(c *Cfg) string {
	if i := strings.IndexByte(c.Name, nameSep); i > 0 {
		return c.Name[:i]
	}
	return c.Name
}

// Add registers a child command with its parent and returns the child.
func (c *Cfg) Add(child *Cfg) *Cfg {
	if child.parent != nil {
		panic("cli: command already has a parent: " + child.Name)
	}
	if child.parent = c; c.cmds == nil {
		c.cmds = make(map[string]*Cfg)
	}
	for name, i := child.Name, 0; ; name = name[i+1:] {
		if i = strings.IndexByte(name, nameSep); i < 0 {
			i = len(name)
		}
		if i == 0 {
			panic("cli: missing command name")
		}
		if _, dup := c.cmds[name[:i]]; dup {
			panic("cli: duplicate command name: " + name[:i])
		}
		if c.cmds[name[:i]] = child; i == len(name) {
			return child
		}
	}
}

// Parse instantiates the requested command and parses the arguments. It returns
// the command, positional arguments, and any UsageError or ErrHelp.
func (c *Cfg) Parse(args []string) (*Cfg, Cmd, []string, error) {
	// Find sub-command
	var err error
	for len(args) > 0 && c.cmds != nil {
		if v := args[0]; isHelp(v) {
			err = ErrHelp
		} else if sub := c.cmds[v]; sub != nil {
			c = sub
		} else if len(v) > 0 {
			err = Errorf("unknown command %q", v)
			break
		}
		args = args[1:]
	}
	if err == nil && len(args) > 0 && isHelp(args[0]) {
		err = ErrHelp
	}

	// Parse options
	cmd := New(c)
	if err == nil && len(args) > 0 {
		fs := NewFlagSet(cmd)
		if err = fs.Parse(args); err != nil && err != ErrHelp {
			err = UsageError(err.Error())
		}
		args = fs.Args()
	}

	// Check positional argument count
	if err != nil {
		args = nil
	} else if c.MinArgs == c.MaxArgs && len(args) != c.MinArgs {
		if c.MinArgs <= 0 {
			err = Error("command does not accept any arguments")
		} else {
			err = Errorf("command requires %d argument(s)", c.MinArgs)
		}
	} else if len(args) < c.MinArgs {
		err = Errorf("command requires at least %d argument(s)", c.MinArgs)
	} else if c.MinArgs < c.MaxArgs && c.MaxArgs < len(args) {
		err = Errorf("command accepts at most %d argument(s)", c.MaxArgs)
	}
	return c, cmd, args, err
}

// Run parses the arguments, runs the requested commands, and terminates the
// process via Exit. If args is nil, os.Args[1:] is used by default.
func (c *Cfg) Run(args ...string) {
	if args == nil {
		args = os.Args[1:]
	}
	c, cmd, args, err := c.Parse(args)
	if err == nil {
		if err = cmd.Main(args); err == nil {
			Exit(0)
			return
		}
	}
	if err == ErrHelp {
		w := newWriter(c)
		defer w.done(os.Stderr, 0)
		w.help()
	} else {
		switch e := err.(type) {
		case UsageError:
			w := newWriter(c)
			defer w.done(os.Stderr, 2)
			w.error(string(e))
		case ExitCode:
			Exit(int(e))
		default:
			verb := "%v"
			if Debug {
				verb = "%+v"
			}
			fmt.Fprintf(os.Stderr, "Error: "+verb+"\n", err)
			Exit(1)
		}
	}
}

// Help returns a buffer containing command help information.
func (c *Cfg) Help() *bytes.Buffer {
	w := newWriter(c)
	w.help()
	return &w.Buffer
}

// Children returns all sub-commands of c sorted by name.
func (c *Cfg) Children() []*Cfg {
	if len(c.cmds) == 0 {
		return nil
	}
	cmds := make([]*Cfg, 0, len(c.cmds))
	for name, child := range c.cmds {
		if name == Name(child) {
			cmds = append(cmds, child)
		}
	}
	sort.Slice(cmds, func(i, j int) bool { return Name(cmds[i]) < Name(cmds[j]) })
	return cmds
}

// fullName returns the fully qualified command name consisting of the prefix,
// primary names of all parents, the primary command name, and any aliases.
func (c *Cfg) fullName(prefix string) string {
	var buf [64]byte
	var walk func(*Cfg)
	b := append(buf[:0], prefix...)
	walk = func(c *Cfg) {
		if c != nil && c.Name != "" {
			walk(c.parent)
			b = append(append(b, ' '), Name(c)...)
		}
	}
	if walk(c.parent); c.Name != "" {
		if b = append(b, ' '); strings.IndexByte(c.Name, nameSep) == -1 {
			b = append(b, c.Name...)
		} else {
			b = append(append(append(b, '{'), c.Name...), '}')
		}
	}
	return strings.TrimSpace(string(b))
}

// nilCmd implements Cmd interface for commands without their own constructor.
type nilCmd Cfg

func (cmd *nilCmd) Main(args []string) error {
	w := newWriter((*Cfg)(cmd))
	defer w.done(os.Stderr, 2)
	if cmd.cmds == nil {
		w.WriteString("Command not implemented\n")
	} else {
		w.WriteString("Specify command:\n")
		w.commands()
	}
	return nil
}
