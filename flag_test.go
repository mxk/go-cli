package cli

import (
	"flag"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFlagSet(t *testing.T) {
	type NoFlag struct{ B byte }
	assert.NotPanics(t, func() { NewFlagSet(new(NoFlag)) })

	type BadType struct {
		B byte `cli:""`
	}
	assert.Panics(t, func() { NewFlagSet(new(BadType)) })

	type S1 struct {
		S1 bool `cli:""`
	}
	type S2 struct {
		S2 bool `cli:""`
	}
	type T struct {
		S1
		ignore   S2
		Set      S2
		NoDesc   bool `cli:""`
		Desc     bool `cli:"Description"`
		Name     bool `cli:"n,"`
		NameDesc bool `cli:"ND,Name description"`
		NoName   bool `cli:"Not a name,"`
		Quote    bool `cli:"Flag, with {value}"`
		NoQuote  bool `cli:"Flag {without value}"`
		Verbose  bool `cli:"V,Verbose"`
	}
	usage := map[string]string{
		"s1":       "",
		"s2":       "",
		"no-desc":  "",
		"desc":     "Description",
		"n":        "",
		"ND":       "Name description",
		"no-name":  "Not a name,",
		"quote":    "Flag, with `value`",
		"no-quote": "Flag {without value}",
		"V":        "Verbose",
	}

	have := T{S1: S1{true}, NoDesc: true}
	fs := NewFlagSet(&have)
	require.NoError(t, fs.Parse(split("-s1=false -s2")))
	require.Equal(t, T{Set: S2{true}, NoDesc: true}, have)

	tf := map[bool]string{false: "false", true: "true"}
	fs.VisitAll(func(f *flag.Flag) {
		u, ok := usage[f.Name]
		if !ok {
			t.Errorf("unknown flag name %q", f.Name)
			return
		}
		assert.Equal(t, u, f.Usage, "%s", f.Name)
		defVal := tf[f.Name == "s1" || f.Name == "no-desc"]
		strVal := tf[f.Name == "s2" || f.Name == "no-desc"]
		assert.Equal(t, defVal, f.DefValue, "%s", f.Name)
		assert.Equal(t, strVal, f.Value.String(), "%s", f.Name)
		delete(usage, f.Name)
	})
	assert.Empty(t, usage)
}

func TestFlagTypes(t *testing.T) {
	ptr := func(x interface{}) interface{} {
		v := reflect.ValueOf(x)
		p := reflect.New(v.Type())
		p.Elem().Set(v)
		return p.Interface()
	}
	type T struct {
		Bool     bool          `cli:""`
		Duration time.Duration `cli:""`
		Float64  float64       `cli:""`
		Int      int           `cli:""`
		Int64    int64         `cli:""`
		String   string        `cli:""`
		Uint     uint          `cli:""`
		Uint64   uint64        `cli:""`
		XY       XY            `cli:""`

		BoolPtr     *bool          `cli:""`
		DurationPtr *time.Duration `cli:""`
		Float64Ptr  *float64       `cli:""`
		IntPtr      *int           `cli:""`
		Int64Ptr    *int64         `cli:""`
		StringPtr   *string        `cli:""`
		UintPtr     *uint          `cli:""`
		Uint64Ptr   *uint64        `cli:""`

		Slice []string          `cli:""`
		Map   map[string]string `cli:""`
	}
	type test struct {
		Name    string
		Default string
		Set     string
		Get     interface{}
		String  string
	}
	tests := []*test{
		{"Bool", "false", "-bool", true, "true"},
		{"Duration", "0s", "-duration=1s", time.Second, "1s"},
		{"Float64", "0", "-float64=0.1", 0.1, "0.1"},
		{"Int", "0", "-int=1", 1, "1"},
		{"Int64", "0", "-int64=-1", int64(-1), "-1"},
		{"String", "", "-string=x", "x", "x"},
		{"Uint", "0", "-uint=1", uint(1), "1"},
		{"Uint64", "0", "-uint64=2", uint64(2), "2"},
		{"XY", "X", "-xy=y", XY('Y'), "Y"},

		{"BoolPtr", "false", "-bool-ptr", ptr(true), "true"},
		{"DurationPtr", "0s", "-duration-ptr=2s", ptr(2 * time.Second), "2s"},
		{"Float64Ptr", "0", "-float64-ptr=0.2", ptr(0.2), "0.2"},
		{"IntPtr", "0", "-int-ptr=2", ptr(2), "2"},
		{"Int64Ptr", "0", "-int64-ptr=-2", ptr(int64(-2)), "-2"},
		{"StringPtr", "", "-string-ptr=y", ptr("y"), "y"},
		{"UintPtr", "0", "-uint-ptr=3", ptr(uint(3)), "3"},
		{"Uint64Ptr", "0", "-uint64-ptr=4", ptr(uint64(4)), "4"},

		{"Slice", "[]", "-slice=a -slice=b", []string{"a", "b"}, "[a b]"},
		{"Map", "{}", "-map=a=1 -map=b=2", map[string]string{"a": "1", "b": "2"}, "{a=1 b=2}"},
	}

	var have, want T
	var args []string
	flagMap := make(map[string]*test, len(tests))
	v := reflect.ValueOf(&want).Elem()
	for _, tc := range tests {
		args = append(args, split(tc.Set)...)
		v.FieldByName(tc.Name).Set(reflect.ValueOf(tc.Get))
		flagMap[flagName(tc.Name)] = tc
	}
	fs := NewFlagSet(&have)
	fs.SetOutput(ioutil.Discard)
	assert.NotPanics(t, func() { fs.PrintDefaults() })
	require.NoError(t, fs.Parse(args))
	require.Equal(t, want, have)

	fs.VisitAll(func(f *flag.Flag) {
		tc := flagMap[f.Name]
		if tc == nil {
			t.Errorf("unknown flag name %q", f.Name)
			return
		}
		assert.Equal(t, tc.Default, f.DefValue, "%+v", tc)
		assert.Equal(t, tc.String, f.Value.String(), "%+v", tc)
		assert.Equal(t, tc.Get, f.Value.(flag.Getter).Get(), "%+v", tc)
		delete(flagMap, f.Name)
	})
	assert.Empty(t, flagMap)
}

func TestFlagName(t *testing.T) {
	tests := []struct{ name, want string }{
		{"", ""},
		{"X", "x"},
		{"x", "x"},
		{"_", ""},

		{"XY", "xy"},
		{"Xy", "xy"},
		{"xY", "x-y"},
		{"xy", "xy"},

		{"_XY", "xy"},
		{"__XY", "xy"},
		{"X_Y", "x-y"},
		{"X_y", "x-y"},
		{"x_Y", "x-y"},
		{"x__y", "x-y"},
		{"XY_", "xy"},
		{"XY__", "xy"},

		{"ABCD", "abcd"},
		{"ABCd", "ab-cd"},
		{"ABcD", "a-bc-d"},
		{"ABcd", "a-bcd"},
		{"AbCD", "ab-cd"},
		{"AbCd", "ab-cd"},
		{"AbcD", "abc-d"},
		{"Abcd", "abcd"},
		{"aBCD", "a-bcd"},
		{"aBCd", "a-b-cd"},
		{"aBcD", "a-bc-d"},
		{"aBcd", "a-bcd"},
		{"abCD", "ab-cd"},
		{"abCd", "ab-cd"},
		{"abcD", "abc-d"},
		{"abcd", "abcd"},

		{"1A", "1-a"},
		{"1a", "1a"},
		{"A1", "a1"},
		{"a1", "a1"},

		{"A1B2", "a1b2"},
		{"A1b2", "a1b2"},
		{"a1B2", "a1-b2"},
		{"a1b2", "a1b2"},
		{"AB1C", "ab1c"},
		{"AB1c", "ab1c"},
		{"Ab1C", "ab1-c"},
		{"Ab1c", "ab1c"},
		{"A1BC", "a1bc"},
		{"A1Bc", "a1-bc"},
		{"A1bC", "a1b-c"},
		{"A1bc", "a1bc"},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, flagName(tc.name), "%+v", tc)
	}
}

type XY byte

func (v XY) String() string {
	if v == 0 {
		return "X"
	}
	return string(v)
}

func (v *XY) Set(s string) error {
	switch s {
	case "x", "X", "y", "Y":
		*v = XY(s[0] &^ 0x20)
		return nil
	}
	return Errorf("invalid enum value %q", s)
}

func (v XY) Get() interface{} { return v }
