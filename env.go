package cli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// SetEnvFields populates struct field values from environment variables using
// "env" field tags as variable names.
func SetEnvFields(structPtr interface{}) error {
	v := reflect.ValueOf(structPtr).Elem()
	t := v.Type()
	for i := t.NumField() - 1; i >= 0; i-- {
		f := t.Field(i)
		key := f.Tag.Get("env")
		if key == "" {
			continue
		}
		val, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		switch f.Type.Kind() {
		case reflect.String:
			v.Field(i).SetString(val)
		case reflect.Bool:
			b := val == ""
			if !b {
				var err error
				if b, err = strconv.ParseBool(val); err != nil {
					return fmt.Errorf("cli: invalid bool value for %q", key)
				}
			}
			v.Field(i).SetBool(b)
		default:
			panic("cli: unsupported field type " + f.Type.String())
		}
	}
	return nil
}

// GetEnvFields extracts environment variable values from struct fields using
// "env" field tags as key names. If all is true, zero values are included in
// the resulting map.
func GetEnvFields(structPtr interface{}, all bool) map[string]string {
	v := reflect.ValueOf(structPtr).Elem()
	t := v.Type()
	m := make(map[string]string)
	for i := t.NumField() - 1; i >= 0; i-- {
		f := t.Field(i)
		key := f.Tag.Get("env")
		if key == "" {
			continue
		}
		var s string
		switch f.Type.Kind() {
		case reflect.String:
			s = v.Field(i).String()
		case reflect.Bool:
			b := v.Field(i).Bool()
			if !b && !all {
				continue
			}
			s = strconv.FormatBool(b)
		default:
			s = fmt.Sprint(v.Field(i).Interface())
		}
		if s != "" || all {
			m[key] = s
		}
	}
	return m
}
