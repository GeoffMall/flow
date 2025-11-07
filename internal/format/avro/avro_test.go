package avro

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/GeoffMall/flow/internal/format"
	"github.com/stretchr/testify/assert"
)

func TestParser_MultipleRecords(t *testing.T) {
	// Open test file with multiple records
	f, err := os.Open("testdata/users.avro")
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
	assert.Equal(t, 30, records[0]["age"])
	assert.Equal(t, true, records[0]["active"])

	// Verify second record
	assert.Equal(t, "Bob", records[1]["name"])
	assert.Equal(t, 25, records[1]["age"])
	assert.Equal(t, false, records[1]["active"])

	// Verify third record
	assert.Equal(t, "Charlie", records[2]["name"])
	assert.Equal(t, 35, records[2]["age"])
	assert.Equal(t, true, records[2]["active"])
}

func TestParser_SingleRecord(t *testing.T) {
	// Open test file with single record
	f, err := os.Open("testdata/single.avro")
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
	assert.Equal(t, 42, records[0]["age"])
	assert.Equal(t, true, records[0]["active"])
}

func TestParser_InvalidFile(t *testing.T) {
	// Try to parse a non-Avro file
	f, err := os.Open("testdata/generate.go")
	assert.NoError(t, err)
	defer f.Close()

	_, err = NewParser(f)
	assert.Error(t, err, "should fail to parse non-Avro file")
}

func TestParser_EarlyReturn(t *testing.T) {
	// Test that ForEach stops when callback returns error
	f, err := os.Open("testdata/users.avro")
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
	assert.Equal(t, "avro", f.Name())
}

func TestFormat_NewParser_Success(t *testing.T) {
	f := &Format{}
	file, err := os.Open("testdata/users.avro")
	assert.NoError(t, err)
	defer file.Close()

	parser, err := f.NewParser(file)
	assert.NoError(t, err)
	assert.NotNil(t, parser)
}

func TestFormat_NewParser_Error(t *testing.T) {
	f := &Format{}
	r := strings.NewReader("not avro data")

	_, err := f.NewParser(r)
	assert.Error(t, err)
}

func TestFormat_NewFormatter_Panics(t *testing.T) {
	f := &Format{}
	var buf bytes.Buffer

	assert.Panics(t, func() {
		_ = f.NewFormatter(&buf, format.FormatterOptions{})
	}, "NewFormatter should panic as Avro write is not supported")
}
