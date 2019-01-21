package build

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var basicFmt = `
version %s
commit  %s
date    %s
go      %s (%s/%s)
`[1:]

var extraFmt = basicFmt + `
abc 123
k   v
`

func TestString(t *testing.T) {
	var (
		gover  = strings.TrimPrefix(runtime.Version(), "go")
		goos   = runtime.GOOS
		goarch = runtime.GOARCH
	)
	Release("", "", "")
	want := fmt.Sprintf(basicFmt, "dev", "-", "-", gover, goos, goarch)
	assert.Equal(t, want, String())

	Release("v1.0.0", "X", "2018-12-15")
	Set("abc", "123")
	SetFrom(map[string]string{"k": "v"})
	want = fmt.Sprintf(extraFmt, "1.0.0", "X", "2018-12-15", gover, goos, goarch)
	assert.Equal(t, want, String())
}
