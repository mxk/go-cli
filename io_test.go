package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	output := interceptWrite(&os.Stdout)
	err := WriteFile("-", func(w io.Writer) error {
		_, err := w.Write([]byte("stdout"))
		return err
	})
	require.NoError(t, err)
	require.Equal(t, "stdout", output())

	tmp, err := ioutil.TempFile("", "cli-write-file-test.")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	tmp.Close()

	want := []byte("hello, world\n")
	err = WriteFile(tmp.Name(), func(w io.Writer) error {
		_, err := w.Write(want)
		return err
	})
	require.NoError(t, err)
	b, err := ioutil.ReadFile(tmp.Name())
	require.NoError(t, err)
	require.Equal(t, want, b)

	err = WriteFile(tmp.Name(), func(w io.Writer) error {
		w.(io.WriteCloser).Close()
		return nil
	})
	errClosed := &os.PathError{Op: "close", Path: tmp.Name(), Err: os.ErrClosed}
	require.Equal(t, errClosed, err)

	errWrite := fmt.Errorf("write error")
	err = WriteFile(tmp.Name(), func(w io.Writer) error {
		w.(io.WriteCloser).Close()
		return errWrite
	})
	require.Equal(t, errWrite, err)
}

func interceptWrite(f **os.File) func() string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	orig := *f
	*f = w
	ch := make(chan string)
	go func() {
		defer close(ch)
		b, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		ch <- string(b)
	}()
	return func() string {
		w.Close()
		select {
		case out := <-ch:
			*f = orig
			return out
		case <-time.After(time.Second):
			panic("timeout")
		}
	}
}
