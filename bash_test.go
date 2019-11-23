package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompgen(t *testing.T) {
	var main Info
	main.Add(&Info{
		Name: "cmd1|c1",
		New:  func() Cmd { return new(cmd1) },
	})
	main.Add(&Info{
		Name: "grp",
	}).Add(&Info{
		Name:    "cmd-2",
		MinArgs: 1,
	})

	want := map[string]*cmdSpec{
		"_": {
			Spec: "-W 'cmd1 grp help'",
		},
		"_cmd1": {
			Name: "cmd1",
			Spec: "-W '-b -d -f -x-z'",
			Refs: []string{"_c1"},
			Args: map[string]string{
				"f":   "-f",
				"d":   "-d",
				"x_z": "-W ''",
			},
		},
		"_grp_": {
			Name: "grp",
			Spec: "-W 'cmd-2 help'",
		},
		"_grp_cmd_2": {
			Name: "cmd_2",
			Spec: "-W '' -o bashdefault",
		},
	}

	have := make(map[string]*cmdSpec)
	newCmdSpec(have, "", &main)
	assert.Equal(t, want, have)
}

type cmd1 struct {
	B  bool   `cli:"bool"`
	F  string `cli:"<file>"`
	D  string `cli:"<dir>"`
	XZ string `cli:"x-z,"`
}

func (*cmd1) Info() *Info              { return nil }
func (*cmd1) Main(args []string) error { return nil }
