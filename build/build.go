// Package build maintains program version and other build information.
package build

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
)

// Global build information.
var (
	Version string
	Commit  string
	Date    string
	Extra   map[string]string
)

// Release sets release information provided by goreleaser.
func Release(version, commit, date string) {
	if version == "" {
		version = "dev"
	} else if version[0] == 'v' {
		version = version[1:]
	}
	if commit == "" {
		commit = "-"
	}
	if date == "" {
		date = "-"
	}
	Version, Commit, Date = version, commit, date
}

// Set sets extra build information.
func Set(k, v string) {
	if Extra == nil {
		Extra = make(map[string]string)
	}
	Extra[k] = v
}

// SetFrom copies extra build information from m.
func SetFrom(m map[string]string) {
	if Extra == nil && len(m) > 0 {
		Extra = make(map[string]string)
	}
	for k, v := range m {
		Extra[k] = v
	}
}

// Write writes build information to stdout.
func Write() error {
	_, err := os.Stdout.WriteString(String())
	return err
}

// String returns complete build information as a string.
func String() string {
	var b strings.Builder
	b.Grow(256)
	fmt.Fprintln(&b, "version", Version)
	fmt.Fprintln(&b, "commit ", Commit)
	fmt.Fprintln(&b, "date   ", Date)
	fmt.Fprintf(&b, "go      %s (%s/%s)\n",
		strings.TrimPrefix(runtime.Version(), "go"),
		runtime.GOOS, runtime.GOARCH)
	if len(Extra) > 0 {
		b.WriteByte('\n')
		w, keys := 0, make([]string, 0, len(Extra))
		for k := range Extra {
			if keys = append(keys, k); len(k) > w {
				w = len(k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "%-*s %s\n", w, k, Extra[k])
		}
	}
	return b.String()
}
