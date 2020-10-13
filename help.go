package cli

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
)

// Writer writes command help information to a buffer.
type Writer struct {
	bytes.Buffer
	*Cfg
}

// newWriter returns a new Writer instance.
func newWriter(c *Cfg) Writer {
	w := Writer{Cfg: c}
	w.Grow(4096)
	return w
}

// Section starts a new help section.
func (w *Writer) Section(name string) {
	b, nl := w.Bytes(), []byte("\n\n")
	if len(b) > 0 && !bytes.HasSuffix(b, nl) {
		if bytes.HasSuffix(b, nl[:1]) {
			nl = nl[1:]
		}
		w.Write(nl)
	}
	if name != "" {
		w.WriteString(name)
		w.WriteString(":\n")
	}
}

// Text writes s to w, removing any indentation and leading/trailing space.
func (w *Writer) Text(s string) {
	w.Section("")
	w.WriteString(strings.TrimSpace(Dedent(s)))
	w.WriteByte('\n')
}

// help writes command help information to w.
func (w *Writer) help() {
	w.usage()
	cmd := New(w.Cfg)
	if h, ok := cmd.(Helper); ok {
		w.WriteByte('\n')
		h.Help(w)
	} else if w.Summary != "" {
		w.WriteByte('\n')
		w.WriteString(w.Summary)
		w.WriteString(".\n")
	}
	if w.cmds != nil {
		w.Section("Commands")
		w.commands()
	} else {
		noOpts := w.Len()
		w.Section("Options")
		fs := NewFlagSet(cmd)
		fs.SetOutput(&w.Buffer)
		ref := w.Len()
		if fs.PrintDefaults(); ref == w.Len() {
			w.Truncate(noOpts)
		}
	}
	w.WriteByte('\n')
}

// error writes command usage error to w.
func (w *Writer) error(msg string) {
	w.WriteString("Error: ")
	w.WriteString(strings.TrimSpace(msg))
	w.WriteByte('\n')
	w.usage()
}

// usage writes command usage summary to w.
func (w *Writer) usage() {
	name := w.fullName(Bin)
	if usage := w.Usage; w.cmds != nil {
		if usage == "" {
			usage = "<command> [options] ..."
		}
		fmt.Fprintf(w, "Usage: %s %s\n", name, usage)
		fmt.Fprintf(w, "       %s <command> help\n", name)
		fmt.Fprintf(w, "       %s help [command]\n", name)
	} else {
		sp := " "
		if usage == "" {
			sp = ""
		}
		fmt.Fprintf(w, "Usage: %s%s%s\n", name, sp, usage)
		fmt.Fprintf(w, "       %s help\n", name)
	}
}

// commands writes a list of all commands with their summaries to w.
func (w *Writer) commands() {
	cmds, maxLen := w.Children(), 0
	for _, c := range cmds {
		if name := Name(c); maxLen < len(name) && !c.Hide {
			maxLen = len(name)
		}
	}
	for _, c := range cmds {
		if !c.Hide {
			if c.Summary == "" {
				fmt.Fprintf(w, "  %s\n", Name(c))
			} else {
				fmt.Fprintf(w, "  %-*s  %s\n", maxLen, Name(c), c.Summary)
			}
		}
	}
}

// done writes the buffer to out and calls Exit.
func (w *Writer) done(out io.Writer, code int) {
	defer Exit(2)
	if p := recover(); p != nil {
		w.WriteString("panic: ")
		fmt.Fprintln(w, p)
		w.WriteByte('\n')
		w.Write(debug.Stack())
		code = 2
	}
	w.WriteTo(out)
	Exit(code)
}

// Dedent removes leading tab characters from each line in s. The first line is
// skipped, the next line containing non-tab characters determines the number of
// tabs to remove.
func Dedent(s string) string {
	n, i := 0, strings.IndexByte(s, '\n')
	for j := i + 1; j < len(s); j++ {
		if c := s[j]; c == '\n' {
			i = j
		} else if c != '\t' {
			n = j - i - 1
			break
		}
	}
	if i == -1 || n == 0 {
		return s
	}
	b := make([]byte, 0, len(s))
	for {
		if i = strings.IndexByte(s, '\n') + 1; i == 0 {
			return string(append(b, s...))
		}
		b, s, i = append(b, s[:i]...), s[i:], 0
		for i < n && i < len(s) && s[i] == '\t' {
			i++
		}
		s = s[i:]
	}
}

// isHelp returns true if s represents a command-line request for help.
func isHelp(s string) bool {
	switch s {
	case "help", "-help", "--help", "-h", "/?":
		return true
	}
	return false
}
