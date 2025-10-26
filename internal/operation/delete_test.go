package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDelete_NonNil(t *testing.T) {
	d := NewDelete([]string{"key1", "key2"})
	assert.NotNil(t, d)
}
