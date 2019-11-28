package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// Stdio returns true if file represents stdin or stdout.
func Stdio(file string) bool { return file == "" || file == "-" }

// WriteFile opens a file for writing and calls fn to write its contents. The
// writer will be os.Stdout if name is empty or "-". The file is created with
// default permissions to match stdout redirect behavior.
func WriteFile(name string, fn func(*os.File) error) (err error) {
	if Stdio(name) {
		return fn(os.Stdout)
	}
	f, err := os.Create(filepath.Clean(name))
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
	}()
	return fn(f)
}

// WriteFileAtomic does the same thing as WriteFile, but uses a temporary file
// for writing, which is either renamed or removed at the end.
func WriteFileAtomic(name string, fn func(*os.File) error) (err error) {
	if Stdio(name) {
		return fn(os.Stdout)
	}
	name = filepath.Clean(name)
	f, err := ioutil.TempFile(filepath.Dir(name), filepath.Base(name)+".*")
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); err == nil {
			if err = e; err == nil {
				err = os.Rename(f.Name(), name)
			}
		}
		if err != nil {
			os.Remove(f.Name())
		}
	}()
	return fn(f)
}
