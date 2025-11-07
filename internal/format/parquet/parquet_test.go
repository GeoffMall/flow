package parquet

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_MultipleRecords(t *testing.T) {
	// Open test file with multiple records
	f, err := os.Open("testdata/users.parquet")
	assert.NoError(t, err)
	defer f.Close()

	parser, err := NewParser(f)
	assert.NoError(t, err)

	// Collect all records
	var records []map[string]any
	err = parser.ForEach(func(doc any) error {
		record, ok := doc.(map[string]any)
		assert.True(t, ok, "document should be a map")
		records = append(records, record)
		return nil
	})
	assert.NoError(t, err)

	// Verify we got 3 records
	assert.Equal(t, 3, len(records))

	// Verify first record
	assert.Equal(t, "Alice", records[0]["name"])
	assert.Equal(t, int32(30), records[0]["age"])
	assert.Equal(t, true, records[0]["active"])

	// Verify second record
	assert.Equal(t, "Bob", records[1]["name"])
	assert.Equal(t, int32(25), records[1]["age"])
	assert.Equal(t, false, records[1]["active"])

	// Verify third record
	assert.Equal(t, "Charlie", records[2]["name"])
	assert.Equal(t, int32(35), records[2]["age"])
	assert.Equal(t, true, records[2]["active"])
}

func TestParser_SingleRecord(t *testing.T) {
	// Open test file with single record
	f, err := os.Open("testdata/single.parquet")
	assert.NoError(t, err)
	defer f.Close()

	parser, err := NewParser(f)
	assert.NoError(t, err)

	// Collect all records
	var records []map[string]any
	err = parser.ForEach(func(doc any) error {
		record, ok := doc.(map[string]any)
		assert.True(t, ok, "document should be a map")
		records = append(records, record)
		return nil
	})
	assert.NoError(t, err)

	// Verify we got 1 record
	assert.Equal(t, 1, len(records))

	// Verify record contents
	assert.Equal(t, "Solo", records[0]["name"])
	assert.Equal(t, int32(42), records[0]["age"])
	assert.Equal(t, true, records[0]["active"])
}

func TestParser_InvalidFile(t *testing.T) {
	// Try to parse a non-Parquet file
	f, err := os.Open("testdata/generate.go")
	assert.NoError(t, err)
	defer f.Close()

	_, err = NewParser(f)
	assert.Error(t, err, "should fail to parse non-Parquet file")
}

func TestParser_StdinError(t *testing.T) {
	// Test that NewParser returns error for stdin (non-seekable input)
	r := strings.NewReader("fake parquet data")
	_, err := NewParser(r)
	assert.Error(t, err, "should fail for non-seekable input")
	assert.Contains(t, err.Error(), "seekable file input", "error should mention seekable requirement")
}

func TestParser_EarlyReturn(t *testing.T) {
	// Test that ForEach stops when callback returns error
	f, err := os.Open("testdata/users.parquet")
	assert.NoError(t, err)
	defer f.Close()

	parser, err := NewParser(f)
	assert.NoError(t, err)

	// Process only first record
	recordCount := 0
	expectedErr := assert.AnError
	err = parser.ForEach(func(doc any) error {
		recordCount++
		return expectedErr // Stop immediately
	})

	assert.Equal(t, expectedErr, err, "should return callback error")
	assert.Equal(t, 1, recordCount, "should have processed exactly 1 record")
}

func TestFormat_Name(t *testing.T) {
	f := &Format{}
	assert.Equal(t, "parquet", f.Name())
}
