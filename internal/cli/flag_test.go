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
		assert.False(t, f.Color, "Color should be false by default")
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
		"--color",
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

		assert.True(t, f.Color, "Color should be true")
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
