package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSum(t *testing.T) {
	assert.Equal(t, 0, Sum())
	assert.Equal(t, 0, Sum(false))
	assert.Equal(t, 1, Sum(true))
	assert.Equal(t, 1, Sum(false, true))
	assert.Equal(t, 2, Sum(true, true))
}
