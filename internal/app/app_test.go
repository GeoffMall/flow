package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/GeoffMall/flow/internal/cli"
	"github.com/stretchr/testify/assert"
)

const fakeJSONPath = "testdata/fake.json"

func Test_run_fakeJSON(t *testing.T) {
	i, closeI, err := openInput(fakeJSONPath)
	assert.NoError(t, err)
	defer closeI()

	var out bytes.Buffer

	opts := &cli.Flags{
		PickPaths: []string{"result[0].name.middle"},
		Compact:   true,
	}
	err = run(i, &out, opts)
	assert.NoError(t, err)

	got := out.String()
	want := `{"result":[{"name":{"middle":"Micah"}}]}` + "\n"
	assert.Equal(t, want, got)
}

func Test_run_EchoJSON_NoOps(t *testing.T) {
	in := strings.NewReader(`{"a":1,"b":2}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		ToFormat: "", // default json
		Compact:  true,
	}
	err := run(in, &out, opts)
	assert.NoError(t, err)

	got := out.String()
	want := `{"a":1,"b":2}` + "\n"
	assert.Equal(t, want, got)
}

func Test_run_PickTwoFields_JSONPretty(t *testing.T) {
	in := strings.NewReader(`{"user":{"name":"alice","id":7,"role":"dev"},"z":0}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		PickPaths: []string{"user.name", "user.id"},
		Compact:   false, // pretty
	}
	err := run(in, &out, opts)
	assert.NoError(t, err)

	got := out.String()
	// Pretty JSON with stable order (indent by printer). We don't enforce key order,
	// but we can check substrings to be robust:
	assert.Contains(t, got, `"name": "alice"`)
	assert.Contains(t, got, `"id": 7`)
}

func Test_run_SetAndDelete(t *testing.T) {
	t.Skip("Skipping for now; needs more robust testing")
	in := strings.NewReader(`{"flags":{"debug":true},"server":{"port":80}}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		SetPairs:    []string{"server.port=8080", "meta.env=\"prod\""},
		DeletePaths: []string{"flags.debug"},
		Compact:     true,
	}
	err := run(in, &out, opts)
	assert.NoError(t, err)

	got := out.String()
	want := `{"meta":{"env":"prod"},"server":{"port":8080},"flags":{}}` + "\n"
	assert.Equal(t, want, got)
}

func Test_run_YAMLIn_JSONOut(t *testing.T) {
	// YAML input with multi-doc; we’ll pick from each and print JSON (default)
	in := strings.NewReader(`---
user:
  name: bob
  id: 1
---
user:
  name: eve
  id: 2
`)
	var out bytes.Buffer

	opts := &cli.Flags{
		PickPaths: []string{"user.id"},
		Compact:   true,
		// ToFormat empty => json
	}
	err := run(in, &out, opts)
	assert.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, 2, len(lines), "expected 2 JSON docs")
	assert.Equal(t, `{"user":{"id":1}}`, lines[0], "first doc")
	assert.Equal(t, `{"user":{"id":2}}`, lines[1], "second doc")
}

func Test_run_JSONIn_YAMLOut(t *testing.T) {
	in := strings.NewReader(`{"a":1,"b":[2,3]}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		ToFormat: "yaml",
	}
	err := run(in, &out, opts)
	assert.NoError(t, err)

	got := out.String()
	// yaml.Encoder produces:
	// a: 1
	// b:
	//   - 2
	//   - 3
	// (with trailing newline)
	assert.Contains(t, got, "a: 1\n")
	assert.Contains(t, got, "b:\n  - 2\n  - 3\n")
}

func TestBuildPipeline_InvalidSet(t *testing.T) {
	opts := &cli.Flags{
		SetPairs: []string{"not-a-pair-with-equals"},
	}
	_, err := buildPipeline(opts)
	assert.Error(t, err, "expected error for invalid set pair")
}
