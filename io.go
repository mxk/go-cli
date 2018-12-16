package cli

import (
	"io"
	"os"
)

// Stdio returns true if file represents stdin or stdout.
func Stdio(file string) bool { return file == "" || file == "-" }

// WriteFile opens a file for writing and calls fn to write its contents. The
// writer will be os.Stdout if name is empty or "-". The file is created with
// default permissions to match stdout redirect behavior.
func WriteFile(name string, fn func(io.Writer) error) (err error) {
	if Stdio(name) {
		return fn(os.Stdout)
	}
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	return fn(f)
}
