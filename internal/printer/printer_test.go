package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_NonNil(t *testing.T) {
	p := New(Options{})
	assert.NotNil(t, p)
}
