package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	strType  string
	boolType bool
	testEnv  struct {
		S1 string   `env:"_S1"`
		S2 strType  `env:"_S2"`
		B1 bool     `env:"_B1"`
		B2 boolType `env:"_B2"`
		B3 boolType `env:"_B3"`
	}
)

func TestSetEnvFields(t *testing.T) {
	env := map[string]string{
		"_S2": "s2",
		"_B2": "",
		"_B3": "false",
	}
	want := testEnv{
		S2: "s2",
		B2: true,
	}
	defer func() {
		for k := range env {
			os.Unsetenv(k)
		}
	}()
	for k, v := range env {
		os.Setenv(k, v)
	}
	var have testEnv
	SetEnvFields(&have)
	assert.Equal(t, want, have)
}

func TestGetEnvFields(t *testing.T) {
	env := testEnv{S2: "s2", B2: true}
	want := map[string]string{
		"_S2": "s2",
		"_B2": "true",
	}
	assert.Equal(t, want, GetEnvFields(&env, false))
	want = map[string]string{
		"_S1": "",
		"_S2": "s2",
		"_B1": "false",
		"_B2": "true",
		"_B3": "false",
	}
	assert.Equal(t, want, GetEnvFields(&env, true))
}
