package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NewFlagSet defines flags using field tags in s, which should be a struct
// pointer.
func NewFlagSet(s interface{}) *flag.FlagSet {
	fs := &flag.FlagSet{Usage: func() {}}
	fs.SetOutput(ioutil.Discard)
	if v := reflect.ValueOf(s); v.Kind() == reflect.Ptr {
		if v = v.Elem(); v.Kind() == reflect.Struct {
			defineFlags(fs, v)
		}
	}
	return fs
}

// defineFlags configures fs using the fields of struct v.
func defineFlags(fs *flag.FlagSet, v reflect.Value) {
	t := v.Type()
	n := v.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup("flag")
		if !ok {
			fv := v.Field(i)
			if fv.Kind() == reflect.Ptr {
				fv = fv.Elem()
			}
			if fv.Kind() == reflect.Struct && fv.CanInterface() {
				defineFlags(fs, fv)
			}
			continue
		}
		j := strings.IndexByte(tag, ',')
		if sp := strings.IndexByte(tag, ' '); 0 <= sp && sp < j {
			j = -1
		}
		name, usage := "", convQuote(tag[j+1:])
		if j > 0 {
			name = tag[:j]
		} else {
			name = strings.ToLower(f.Name)
		}
		switch p := v.Field(i).Addr().Interface().(type) {
		case *bool:
			fs.BoolVar(p, name, *p, usage)
		case *time.Duration:
			fs.DurationVar(p, name, *p, usage)
		case *float64:
			fs.Float64Var(p, name, *p, usage)
		case *int:
			fs.IntVar(p, name, *p, usage)
		case *int64:
			fs.Int64Var(p, name, *p, usage)
		case *string:
			fs.StringVar(p, name, *p, usage)
		case *uint:
			fs.UintVar(p, name, *p, usage)
		case *uint64:
			fs.Uint64Var(p, name, *p, usage)
		case flag.Value:
			fs.Var(p, name, usage)
		case **bool:
			fs.Var(boolPtr{p}, name, usage)
		case **time.Duration:
			fs.Var(durPtr{p}, name, usage)
		case **float64:
			fs.Var(f64Ptr{p}, name, usage)
		case **int:
			fs.Var(intPtr{p}, name, usage)
		case **int64:
			fs.Var(i64Ptr{p}, name, usage)
		case **string:
			fs.Var(strPtr{p}, name, usage)
		case **uint:
			fs.Var(uintPtr{p}, name, usage)
		case **uint64:
			fs.Var(u64Ptr{p}, name, usage)
		case *[]string:
			fs.Var(strSlice{p}, name, usage)
		case *map[string]string:
			fs.Var(strMap{p}, name, usage)
		default:
			panic("cli: unsupported flag type: " + f.Type.String())
		}
	}
}

// convQuote converts "<name>" to "`name`" in usage strings. This format is
// easier to use in struct field tags.
func convQuote(usage string) string {
	i := -1
	for j := range usage {
		switch usage[j] {
		case '<':
			if i == -1 {
				i = j
			}
		case '>':
			if i != -1 {
				return usage[:i] + "`" + usage[i+1:j] + "`" + usage[j+1:]
			}
		case '\t', ' ':
			i = -1
		}
	}
	return usage
}

// boolPtr implements flag.Value for *bool flags.
type boolPtr struct{ v **bool }

func (p boolPtr) String() string {
	var v bool
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return strconv.FormatBool(v)
}

func (p boolPtr) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*p.v = &v
	return err
}

func (p boolPtr) Get() interface{} { return *p.v }

func (p boolPtr) IsBoolFlag() bool { return true }

// durPtr implements flag.Value for *time.Duration flags.
type durPtr struct{ v **time.Duration }

func (p durPtr) String() string {
	var v time.Duration
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return v.String()
}

func (p durPtr) Set(s string) error {
	v, err := time.ParseDuration(s)
	*p.v = &v
	return err
}

func (p durPtr) Get() interface{} { return *p.v }

// f64Ptr implements flag.Value for *float64 flags.
type f64Ptr struct{ v **float64 }

func (p f64Ptr) String() string {
	var v float64
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func (p f64Ptr) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*p.v = &v
	return err
}

func (p f64Ptr) Get() interface{} { return *p.v }

// intPtr implements flag.Value for *int flags.
type intPtr struct{ v **int }

func (p intPtr) String() string {
	var v int
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return strconv.Itoa(v)
}

func (p intPtr) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	i := int(v)
	*p.v = &i
	return err
}

func (p intPtr) Get() interface{} { return *p.v }

// i64Ptr implements flag.Value for *int flags.
type i64Ptr struct{ v **int64 }

func (p i64Ptr) String() string {
	var v int64
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return strconv.FormatInt(v, 10)
}

func (p i64Ptr) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*p.v = &v
	return err
}

func (p i64Ptr) Get() interface{} { return *p.v }

// strPtr implements flag.Value for *string flags.
type strPtr struct{ v **string }

func (p strPtr) String() string {
	var v string
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return v
}

func (p strPtr) Set(s string) error {
	*p.v = &s
	return nil
}

func (p strPtr) Get() interface{} { return *p.v }

// uintPtr implements flag.Value for *uint flags.
type uintPtr struct{ v **uint }

func (p uintPtr) String() string {
	var v uint
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return strconv.FormatUint(uint64(v), 10)
}

func (p uintPtr) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	u := uint(v)
	*p.v = &u
	return err
}

func (p uintPtr) Get() interface{} { return *p.v }

// u64Ptr implements flag.Value for *uint64 flags.
type u64Ptr struct{ v **uint64 }

func (p u64Ptr) String() string {
	var v uint64
	if p.v != nil && *p.v != nil {
		v = **p.v
	}
	return strconv.FormatUint(v, 10)
}

func (p u64Ptr) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*p.v = &v
	return err
}

func (p u64Ptr) Get() interface{} { return *p.v }

// strSlice implements flag.Value for []string flags.
type strSlice struct{ v *[]string }

func (p strSlice) String() string {
	var v []string
	if p.v != nil {
		v = *p.v
	}
	return fmt.Sprint(v)
}

func (p strSlice) Set(s string) error {
	*p.v = append(*p.v, s)
	return nil
}

func (p strSlice) Get() interface{} { return *p.v }

// strMap implements flag.Value for map[string]string flags.
type strMap struct{ v *map[string]string }

func (p strMap) String() string {
	var v map[string]string
	if p.v != nil {
		v = *p.v
	}
	var sb strings.Builder
	sb.Grow(2 + 16*len(v))
	sb.WriteByte('{')
	if len(v) > 0 {
		keys := make([]string, 0, len(v))
		for k := range *p.v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(v[k])
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func (p strMap) Set(s string) error {
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return fmt.Errorf("cli: missing '=' in %q", s)
	}
	if *p.v == nil {
		*p.v = make(map[string]string)
	}
	(*p.v)[s[:i]] = s[i+1:]
	return nil
}

func (p strMap) Get() interface{} { return *p.v }
