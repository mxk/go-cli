package cli

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	output := interceptWrite(&os.Stdout)
	require.NoError(t, WriteFile("-", func(f *os.File) error {
		_, err := f.Write([]byte("stdout"))
		return err
	}))
	require.Equal(t, "stdout", output())

	out := tmpFile(t)
	defer os.Remove(out)

	want := []byte("hello, world")
	require.NoError(t, WriteFile(out, func(f *os.File) error {
		_, err := f.Write(want)
		return err
	}))
	requireFileEqual(t, out, want)

	errClosed := &os.PathError{Op: "close", Path: out, Err: os.ErrClosed}
	require.Equal(t, errClosed, WriteFile(out, func(f *os.File) error {
		f.Close()
		return nil
	}))

	require.Equal(t, assert.AnError, WriteFile(out, func(f *os.File) error {
		f.Close()
		return assert.AnError
	}))
}

func TestWriteFileAtomic(t *testing.T) {
	output := interceptWrite(&os.Stdout)
	require.NoError(t, WriteFileAtomic("", func(f *os.File) error {
		_, err := f.Write([]byte("stdout"))
		return err
	}))
	require.Equal(t, "stdout", output())

	out := tmpFile(t)
	defer os.Remove(out)
	want := []byte("old file")
	require.NoError(t, ioutil.WriteFile(out, want, 0666))

	var tmp string
	require.Equal(t, assert.AnError, WriteFileAtomic(out, func(f *os.File) error {
		tmp = f.Name()
		require.FileExists(t, tmp)
		f.Write([]byte("ignore"))
		return assert.AnError
	}))
	_, err := os.Stat(tmp)
	require.True(t, os.IsNotExist(err), "%v", err)
	requireFileEqual(t, out, want)

	want = []byte("new file")
	require.NoError(t, WriteFileAtomic(out, func(f *os.File) error {
		tmp = f.Name()
		_, err := f.Write(want)
		return err
	}))
	_, err = os.Stat(tmp)
	require.True(t, os.IsNotExist(err), "%v", err)
	requireFileEqual(t, out, want)
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

func tmpFile(t *testing.T) string {
	f, err := ioutil.TempFile("", t.Name()+".*")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func requireFileEqual(t *testing.T, name string, want []byte) {
	have, err := ioutil.ReadFile(name)
	require.NoError(t, err)
	require.Equal(t, want, have)
}
