package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugFromEnv(t *testing.T) {
	tests := []*struct {
		unset bool
		value string
		debug bool
		spec  string
	}{
		{unset: true},
		{debug: true},
		{value: "0"},
		{value: "true", debug: true},
		{value: "x", debug: true, spec: "x"},
	}
	var key = t.Name()
	defer os.Unsetenv(key)
	defer func() { Debug, DebugSpec = false, "" }()
	for _, tc := range tests {
		if Debug, DebugSpec = false, ""; tc.unset {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, tc.value)
		}
		DebugFromEnv(key)
		assert.Equal(t, tc.debug, Debug, "%+v", tc)
		assert.Equal(t, tc.spec, DebugSpec, "%+v", tc)
	}
}

func TestSum(t *testing.T) {
	assert.Equal(t, 0, Sum())
	assert.Equal(t, 0, Sum(false))
	assert.Equal(t, 1, Sum(true))
	assert.Equal(t, 1, Sum(false, true))
	assert.Equal(t, 2, Sum(true, true))
}
