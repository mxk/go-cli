package template

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"text/template"
)

// Show is a special -template argument that causes the default template to be
// written without executing it.
const Show = "show"

// Cmd provides common functionality for implementing commands that render data
// with text/template package. Default should be set to the default template
// contents.
type Cmd struct {
	JSON     bool   `flag:"Write raw data in JSON format"`
	Out      string `flag:"Output <file> name"`
	Template string `flag:"Custom template <file> (use '-' for stdin or 'show' to show default)"`

	Default string
}

// Write writes rendered data to the output file.
func (c *Cmd) Write(data interface{}) error {
	var b bytes.Buffer
	if err := c.Execute(&b, data); err != nil {
		return err
	}
	if c.Out == "" || c.Out == "-" {
		_, err := b.WriteTo(os.Stdout)
		return err
	}
	return ioutil.WriteFile(c.Out, b.Bytes(), 0666)
}

// Execute applies current template to data and writes the output to w.
func (c *Cmd) Execute(w io.Writer, data interface{}) error {
	if c.JSON {
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "\t")
		return enc.Encode(data)
	}
	var t *template.Template
	var err error
	switch c.Template {
	case "":
		t, err = template.New("").Parse(c.Default)
	case "-":
		b, ioerr := ioutil.ReadAll(io.LimitReader(os.Stdin, 8*1024*1024))
		if ioerr != nil {
			return ioerr
		}
		t, err = template.New("").Parse(string(b))
	case Show:
		_, err = io.WriteString(w, c.Default)
		return err
	default:
		t, err = template.ParseFiles(c.Template)
	}
	if err == nil {
		err = t.Execute(w, data)
	}
	return err
}
