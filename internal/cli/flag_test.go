package cli

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags_Defaults(t *testing.T) {
	resetGlobalFlags()

	withArgs(t, []string{}, func() {
		f := ParseFlags()

		assert.Equal(t, "", f.InputFile, "InputFile should be empty by default")
		assert.Equal(t, "", f.OutputFile, "OutputFile should be empty by default")
		assert.Equal(t, 0, len(f.PickPaths), "PickPaths should be empty by default")
		assert.Equal(t, 0, len(f.SetPairs), "SetPairs should be empty by default")
		assert.Equal(t, 0, len(f.DeletePaths), "DeletePaths should be empty by default")
		assert.False(t, f.NoColor, "NoColor should be false by default")
		assert.False(t, f.Compact, "Compact should be false by default")
		assert.Equal(t, "", f.ToFormat, "ToFormat should be empty by default")
	})
}

func TestParseFlags_AllFlags(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--in", "in.json",
		"--out", "out.yaml",
		"--pick", "user.name",
		"--pick", "user.id",
		"--set", "server.port=8080",
		"--set", "debug=true",
		"--delete", "server.secret",
		"--no-color",
		"--compact",
		"--to", "yaml",
	}

	withArgs(t, args, func() {
		f := ParseFlags()

		assert.Equal(t, "in.json", f.InputFile, "InputFile should match")
		assert.Equal(t, "out.yaml", f.OutputFile, "OutputFile should match")

		wantPick := []string{"user.name", "user.id"}
		assert.Equal(t, wantPick, f.PickPaths, "PickPaths should match")

		wantSet := []string{"server.port=8080", "debug=true"}
		assert.Equal(t, wantSet, f.SetPairs, "SetPairs should match")

		wantDel := []string{"server.secret"}
		assert.Equal(t, wantDel, f.DeletePaths, "DeletePaths should match")

		assert.True(t, f.NoColor, "NoColor should be true")
		assert.True(t, f.Compact, "Compact should be true")
		assert.Equal(t, "yaml", f.ToFormat, "ToFormat should match")
	})
}

// withArgs temporarily sets os.Args for a test.
func withArgs(t *testing.T, args []string, fn func()) {
	t.Helper()

	origArgs := os.Args

	defer func() { os.Args = origArgs }()

	os.Args = append([]string{origArgs[0]}, args...)

	fn()
}

// resetGlobalFlags resets the package-level flag.CommandLine so tests don't interfere.
func resetGlobalFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestParseFlags_WithWhere(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--where", "name=Alice",
		"--where", "age=30",
	}

	withArgs(t, args, func() {
		f := ParseFlags()

		wantWhere := []string{"name=Alice", "age=30"}
		assert.Equal(t, wantWhere, f.WherePairs, "WherePairs should match")
	})
}

func TestParseFlags_WithInputDir(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--in-dir", "./testdata",
		"--from", "avro",
	}

	withArgs(t, args, func() {
		f := ParseFlags()

		assert.Equal(t, "./testdata", f.InputDir, "InputDir should match")
		assert.Equal(t, "avro", f.FromFormat, "FromFormat should match")
	})
}

func TestParseFlags_AvroFormat(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--from", "avro",
	}

	withArgs(t, args, func() {
		f := ParseFlags()
		assert.Equal(t, "avro", f.FromFormat)
	})
}

func TestParseFlags_ParquetFormat(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--from", "parquet",
	}

	withArgs(t, args, func() {
		f := ParseFlags()
		assert.Equal(t, "parquet", f.FromFormat)
	})
}

func TestParseFlags_PreserveHierarchy(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--preserve-hierarchy",
	}

	withArgs(t, args, func() {
		f := ParseFlags()
		assert.True(t, f.PreserveHierarchy)
	})
}

func TestParseFlags_YAMLFormat(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--from", "yaml",
		"--to", "json",
	}

	withArgs(t, args, func() {
		f := ParseFlags()
		assert.Equal(t, "yaml", f.FromFormat)
		assert.Equal(t, "json", f.ToFormat)
	})
}

func TestParseFlags_MultipleOperations(t *testing.T) {
	resetGlobalFlags()

	args := []string{
		"--pick", "name",
		"--set", "status=active",
		"--delete", "debug",
		"--where", "type=user",
	}

	withArgs(t, args, func() {
		f := ParseFlags()
		assert.Equal(t, []string{"name"}, f.PickPaths)
		assert.Equal(t, []string{"status=active"}, f.SetPairs)
		assert.Equal(t, []string{"debug"}, f.DeletePaths)
		assert.Equal(t, []string{"type=user"}, f.WherePairs)
	})
}

func TestMultiStringFlag_String(t *testing.T) {
	msf := multiStringFlag{"a", "b", "c"}
	assert.Equal(t, "a, b, c", msf.String())
}

func TestMultiStringFlag_Set(t *testing.T) {
	var msf multiStringFlag
	err := msf.Set("value1")
	assert.NoError(t, err)
	err = msf.Set("value2")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value1", "value2"}, []string(msf))
}
