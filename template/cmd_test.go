package template

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	var b strings.Builder
	require.NoError(t, (&Cmd{JSON: true}).Execute(&b, 1))
	assert.Equal(t, "1\n", b.String())
	b.Reset()

	cmd := Cmd{Default: "Value: {{.}}"}
	require.NoError(t, cmd.Execute(&b, 2))
	assert.Equal(t, "Value: 2", b.String())
	b.Reset()

	cmd.Template = Show
	require.NoError(t, cmd.Execute(&b, nil))
	assert.Equal(t, "Value: {{.}}", b.String())
	b.Reset()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = orig }()
	go func() {
		defer w.Close()
		w.WriteString("Custom: {{.}}")
	}()

	cmd.Template = "-"
	require.NoError(t, cmd.Execute(&b, true))
	assert.Equal(t, "Custom: true", b.String())
	b.Reset()
}
