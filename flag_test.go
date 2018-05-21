package cli

import (
	"flag"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Enum byte

func (e Enum) String() string { return string(e) }

func (e *Enum) Set(s string) error {
	switch s {
	case "x", "y":
		*e = Enum(s[0])
		return nil
	}
	return Errorf("invalid enum value %q", s)
}

func TestNewFlagSet(t *testing.T) {
	type Good struct{ B byte }
	assert.NotPanics(t, func() { NewFlagSet(new(Good)) })

	type Bad struct {
		B byte `flag:""`
	}
	assert.Panics(t, func() { NewFlagSet(new(Bad)) })

	type B1 struct {
		B1 bool `flag:""`
	}
	type B2 struct {
		B2 bool `flag:""`
	}
	type T struct {
		B1
		B2  *B2
		Nil *B2
		B   bool          `flag:"b3,"`
		D   time.Duration `flag:""`
		F64 float64       `flag:"f,Float option description"`
		I   int           "flag:\",Set `int` value, if you want\""
		I64 int64         `flag:""`
		S   string        `flag:""`
		U   uint          `flag:""`
		U64 uint64        `flag:""`
		V   Enum          `flag:""`
	}
	names := split("b1 b2 b3 d f i i64 s u u64 v")

	v := T{B2: new(B2), B: true, S: "default", V: Enum('x')}
	want := v
	fs := NewFlagSet(&v)
	require.NoError(t, fs.Parse(split("-b1 -b3=false -i=1 -v y")))
	want.B1.B1 = true
	want.B = false
	want.I = 1
	want.V = Enum('y')
	require.Equal(t, want, v)

	var actual []string
	fs.VisitAll(func(f *flag.Flag) {
		switch f.Name {
		case "b1", "b2", "b3":
			assert.Equal(t, "", f.Usage)
			assert.Equal(t, strconv.FormatBool(f.Name == "b1"), f.Value.String())
			assert.Equal(t, strconv.FormatBool(f.Name == "b3"), f.DefValue)
		case "d":
			assert.Equal(t, "", f.Usage)
			assert.Equal(t, "0s", f.Value.String())
			assert.Equal(t, "0s", f.DefValue)
		case "f":
			assert.Equal(t, "Float option description", f.Usage)
			assert.Equal(t, "0", f.Value.String())
			assert.Equal(t, "0", f.DefValue)
		case "i":
			assert.Equal(t, "Set `int` value, if you want", f.Usage)
			assert.Equal(t, "1", f.Value.String())
			assert.Equal(t, "0", f.DefValue)
		case "i64":
			assert.Equal(t, "", f.Usage)
			assert.Equal(t, "0", f.Value.String())
			assert.Equal(t, "0", f.DefValue)
		case "s":
			assert.Equal(t, "", f.Usage)
			assert.Equal(t, "default", f.Value.String())
			assert.Equal(t, "default", f.DefValue)
		case "u":
			assert.Equal(t, "", f.Usage)
			assert.Equal(t, "0", f.Value.String())
			assert.Equal(t, "0", f.DefValue)
		case "u64":
			assert.Equal(t, "", f.Usage)
			assert.Equal(t, "0", f.Value.String())
			assert.Equal(t, "0", f.DefValue)
		case "v":
			assert.Equal(t, "", f.Usage)
			assert.Equal(t, "y", f.Value.String())
			assert.Equal(t, "x", f.DefValue)
		default:
			t.Fatalf("unexpected flag name %q", f.Name)
		}
		actual = append(actual, f.Name)
	})
	assert.Equal(t, names, actual)
}

func TestPtrValues(t *testing.T) {
	type T struct {
		B   *bool          `flag:"Test"`
		D   *time.Duration `flag:""`
		F   *float64       `flag:""`
		I   *int           `flag:""`
		I64 *int64         `flag:""`
		S   *string        `flag:""`
		U   *uint          `flag:""`
		U64 *uint64        `flag:""`
	}
	v := T{}
	require.NoError(t, NewFlagSet(&v).Parse(nil))
	require.Equal(t, T{}, v)

	cl := "-b=false -d=0s -f=0 -i=0 -i64=0 -s= -u=0 -u64=0"
	fs := NewFlagSet(&v)
	require.NoError(t, fs.Parse(split(cl)))
	b := false
	d := time.Duration(0)
	f := float64(0)
	i := int(0)
	i64 := int64(0)
	s := ""
	u := uint(0)
	u64 := uint64(0)
	want := T{&b, &d, &f, &i, &i64, &s, &u, &u64}
	require.Equal(t, want, v)

	fs.VisitAll(func(f *flag.Flag) {
		switch g := f.Value.(flag.Getter).Get(); f.Name {
		case "b":
			assert.Equal(t, v.B, g.(*bool))
		case "d":
			assert.Equal(t, v.D, g.(*time.Duration))
		case "f":
			assert.Equal(t, v.F, g.(*float64))
		case "i":
			assert.Equal(t, v.I, g.(*int))
		case "i64":
			assert.Equal(t, v.I64, g.(*int64))
		case "s":
			assert.Equal(t, v.S, g.(*string))
		case "u":
			assert.Equal(t, v.U, g.(*uint))
		case "u64":
			assert.Equal(t, v.U64, g.(*uint64))
		default:
			t.Fatalf("unexpected flag name %q", f.Name)
		}
	})

	cl = "-b -d=1s -f=1.0 -i=1 -i64=0x1 -s=xyz -u=1 -u64=0x1"
	require.NoError(t, NewFlagSet(&v).Parse(split(cl)))
	require.NotEqual(t, want, v)

	cl = "-b=false -d=0s -f=0 -i=0 -i64=0 -s= -u=0 -u64=0"
	require.NoError(t, NewFlagSet(&v).Parse(split(cl)))
	require.Equal(t, want, v)
}
