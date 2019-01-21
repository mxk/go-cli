package template

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/mxk/cloudcover/x/cli"
)

// Show is a special -template argument that causes the default template to be
// written without executing it.
const Show = "show"

// FuncMap is an alias for standard FuncMap type.
type FuncMap = template.FuncMap

// Cmd provides common functionality for implementing commands that render data
// with text/template package. Default should be set to the default template
// contents.
type Cmd struct {
	JSON     bool   `flag:"Write raw data in JSON format"`
	Out      string `flag:"Output <file>"`
	Template string `flag:"Custom template <file> (use '-' for stdin or 'show' to show default)"`

	Default string
	FuncMap FuncMap
	Options []string
}

// Write writes rendered data to the output file.
func (c *Cmd) Write(data interface{}) error {
	var b bytes.Buffer
	if err := c.Execute(&b, data); err != nil {
		return err
	}
	return cli.WriteFile(c.Out, func(w io.Writer) error {
		_, err := b.WriteTo(w)
		return err
	})
}

// Execute applies current template to data and writes the output to w.
func (c *Cmd) Execute(w io.Writer, data interface{}) error {
	if c.JSON {
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "\t")
		return enc.Encode(data)
	}
	t := template.New("").Funcs(c.FuncMap).Option(c.Options...)
	var err error
	switch c.Template {
	case "":
		_, err = t.Parse(c.Default)
	case Show:
		_, err = io.WriteString(w, c.Default)
		return err
	default:
		var b []byte
		if c.Template == "-" {
			b, err = ioutil.ReadAll(io.LimitReader(os.Stdin, 8*1024*1024))
		} else {
			b, err = ioutil.ReadFile(c.Template)
		}
		if err == nil {
			_, err = t.Parse(string(b))
		}
	}
	if err == nil {
		err = t.Execute(w, data)
	}
	return err
}
