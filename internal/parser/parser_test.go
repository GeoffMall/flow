package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_NoErrOnOsStdin(t *testing.T) {
	_, err := New(os.Stdin)
	assert.NoError(t, err)
}
