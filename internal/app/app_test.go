package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/GeoffMall/flow/internal/cli"
	"github.com/stretchr/testify/assert"
)

const fakeJSONPath = "testdata/fake.json"

// helper functions for common test patterns

func runTest(t *testing.T, input string, opts *cli.Flags) (string, error) {
	in := strings.NewReader(input)
	var out bytes.Buffer

	err := run(in, &out, opts)
	return out.String(), err
}

func assertRunSucceeds(t *testing.T, input string, opts *cli.Flags, expectedContent string) {
	got, err := runTest(t, input, opts)
	assert.NoError(t, err)
	assert.Contains(t, got, expectedContent)
}

func assertRunFails(t *testing.T, input string, opts *cli.Flags) {
	_, err := runTest(t, input, opts)
	assert.Error(t, err)
}

func Test_run_fakeJSON(t *testing.T) {
	t.Run("jq-like", func(t *testing.T) {
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
		want := `"Micah"` + "\n" // Just the value in jq-like mode
		assert.Equal(t, want, got)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		i, closeI, err := openInput(fakeJSONPath)
		assert.NoError(t, err)
		defer closeI()

		var out bytes.Buffer

		opts := &cli.Flags{
			PickPaths:         []string{"result[0].name.middle"},
			Compact:           true,
			PreserveHierarchy: true,
		}
		err = run(i, &out, opts)
		assert.NoError(t, err)

		got := out.String()
		want := `{"result":[{"name":{"middle":"Micah"}}]}` + "\n"
		assert.Equal(t, want, got)
	})
}

func Test_run_EchoJSON_NoOps(t *testing.T) {
	opts := &cli.Flags{
		ToFormat: "", // default json
		Compact:  true,
	}
	got, err := runTest(t, `{"a":1,"b":2}`, opts)
	assert.NoError(t, err)
	assert.Equal(t, `{"a":1,"b":2}`+"\n", got)
}

func Test_run_PickTwoFields_JSONPretty(t *testing.T) {
	input := `{"user":{"name":"alice","id":7,"role":"dev"},"z":0}`

	t.Run("jq-like", func(t *testing.T) {
		in := strings.NewReader(input)
		var out bytes.Buffer

		opts := &cli.Flags{
			PickPaths: []string{"user.name", "user.id"},
			Compact:   false, // pretty
		}
		err := run(in, &out, opts)
		assert.NoError(t, err)

		got := out.String()
		// In jq-like mode, multiple paths are flattened to root level
		assert.Contains(t, got, `"name": "alice"`)
		assert.Contains(t, got, `"id": 7`)
		// Should NOT contain "user" key in jq-like mode
		assert.NotContains(t, got, `"user":`)
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		in := strings.NewReader(input)
		var out bytes.Buffer

		opts := &cli.Flags{
			PickPaths:         []string{"user.name", "user.id"},
			Compact:           false, // pretty
			PreserveHierarchy: true,
		}
		err := run(in, &out, opts)
		assert.NoError(t, err)

		got := out.String()
		// In preserve-hierarchy mode, maintains structure
		assert.Contains(t, got, `"user":`)
		assert.Contains(t, got, `"name": "alice"`)
		assert.Contains(t, got, `"id": 7`)
	})
}

func Test_run_SetAndDelete(t *testing.T) {
	opts := &cli.Flags{
		SetPairs:    []string{"server.port=8080", "meta.env=\"prod\""},
		DeletePaths: []string{"flags.debug"},
		Compact:     true,
	}
	got, err := runTest(t, `{"flags":{"debug":true},"server":{"port":80}}`, opts)
	assert.NoError(t, err)

	// Check that the operations were applied correctly
	assert.Contains(t, got, `"server":{"port":8080}`)
	assert.Contains(t, got, `"meta":{"env":"prod"}`)
	assert.Contains(t, got, `"flags":{}`) // empty flags object after debug deletion
}

func Test_run_YAMLIn_JSONOut(t *testing.T) {
	yamlInput := `---
user:
  name: bob
  id: 1
---
user:
  name: eve
  id: 2
`

	t.Run("jq-like", func(t *testing.T) {
		// YAML input with multi-doc; we'll pick from each and print JSON (default)
		in := strings.NewReader(yamlInput)
		var out bytes.Buffer

		opts := &cli.Flags{
			PickPaths:  []string{"user.id"},
			Compact:    true,
			FromFormat: "yaml",
			// ToFormat empty => json
		}
		err := run(in, &out, opts)
		assert.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		assert.Equal(t, 2, len(lines), "expected 2 JSON docs")
		assert.Equal(t, `1`, lines[0], "first doc - just the value")
		assert.Equal(t, `2`, lines[1], "second doc - just the value")
	})

	t.Run("preserve-hierarchy", func(t *testing.T) {
		in := strings.NewReader(yamlInput)
		var out bytes.Buffer

		opts := &cli.Flags{
			PickPaths:         []string{"user.id"},
			Compact:           true,
			FromFormat:        "yaml",
			PreserveHierarchy: true,
			// ToFormat empty => json
		}
		err := run(in, &out, opts)
		assert.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		assert.Equal(t, 2, len(lines), "expected 2 JSON docs")
		assert.Equal(t, `{"user":{"id":1}}`, lines[0], "first doc")
		assert.Equal(t, `{"user":{"id":2}}`, lines[1], "second doc")
	})
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

func Test_run_InvalidJSON(t *testing.T) {
	opts := &cli.Flags{Compact: true}
	assertRunFails(t, `{"invalid": json syntax}`, opts)
}

func Test_run_EmptyInput(t *testing.T) {
	opts := &cli.Flags{Compact: true}
	_, err := runTest(t, "", opts)
	assert.Error(t, err, "empty input should cause EOF error")
}

func Test_run_InvalidYAML(t *testing.T) {
	opts := &cli.Flags{Compact: true}
	assertRunFails(t, `invalid: yaml: content: - with bad syntax`, opts)
}

//nolint:funlen // Table-driven test covering comprehensive JSON data types
func Test_run_JSONDataTypes_Comprehensive(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		compact   bool
		checkFunc func(t *testing.T, got string)
	}{
		{
			name:    "null_values",
			input:   `{"key":null,"other":"value"}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"key":null`)
				assert.Contains(t, got, `"other":"value"`)
			},
		},
		{
			name:    "float_numbers",
			input:   `{"pi":3.14159,"negative":-2.5}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"pi":3.14159`)
				assert.Contains(t, got, `"negative":-2.5`)
			},
		},
		{
			name:    "scientific_notation",
			input:   `{"big":1e10,"small":1e-5}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"big":10000000000`)
				assert.Contains(t, got, `"small":0.00001`)
			},
		},
		{
			name:    "unicode_strings",
			input:   `{"name":"José","emoji":"🚀"}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"name":"José"`)
				assert.Contains(t, got, `"emoji":"🚀"`)
			},
		},
		{
			name:    "special_chars",
			input:   `{"quote":"He said \"hello\"","newline":"line1\nline2"}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"quote":"He said \"hello\""`)
				assert.Contains(t, got, `"newline":"line1\nline2"`)
			},
		},
		{
			name:    "empty_structures",
			input:   `{"empty_obj":{},"empty_arr":[]}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"empty_obj":{}`)
				assert.Contains(t, got, `"empty_arr":[]`)
			},
		},
		{
			name:    "mixed_array",
			input:   `[1,"string",true,null,{"nested":"value"}]`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				// Arrays are streamed as separate documents
				lines := strings.Split(strings.TrimSpace(got), "\n")
				assert.Equal(t, 5, len(lines))
				assert.Equal(t, "1", lines[0])
				assert.Equal(t, `"string"`, lines[1])
				assert.Equal(t, "true", lines[2])
				assert.Equal(t, "null", lines[3])
				assert.Contains(t, lines[4], `"nested":"value"`)
			},
		},
		{
			name:    "deeply_nested",
			input:   `{"a":{"b":{"c":{"d":"deep"}}}}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"a":{"b":{"c":{"d":"deep"}}}`)
			},
		},
		{
			name:    "array_of_objects",
			input:   `[{"id":1,"name":"first"},{"id":2,"name":"second"}]`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				// Arrays are streamed as separate documents
				lines := strings.Split(strings.TrimSpace(got), "\n")
				assert.Equal(t, 2, len(lines))
				assert.Contains(t, lines[0], `"id":1`)
				assert.Contains(t, lines[0], `"name":"first"`)
				assert.Contains(t, lines[1], `"id":2`)
				assert.Contains(t, lines[1], `"name":"second"`)
			},
		},
		{
			name:    "boolean_values",
			input:   `{"enabled":true,"disabled":false}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"enabled":true`)
				assert.Contains(t, got, `"disabled":false`)
			},
		},
		{
			name:    "zero_values",
			input:   `{"zero":0,"false":false,"empty":""}`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"zero":0`)
				assert.Contains(t, got, `"false":false`)
				assert.Contains(t, got, `"empty":""`)
			},
		},
		{
			name:    "whitespace_variations",
			input:   `{"key" :  "value" , "next" :123 }`,
			compact: true,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"key":"value"`)
				assert.Contains(t, got, `"next":123`)
			},
		},
		{
			name:    "pretty_format",
			input:   `{"compact":false}`,
			compact: false,
			checkFunc: func(t *testing.T, got string) {
				assert.Contains(t, got, `"compact": false`)
				assert.Contains(t, got, "\n  ") // indented
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			var out bytes.Buffer

			opts := &cli.Flags{
				Compact: tt.compact,
			}
			err := run(in, &out, opts)
			assert.NoError(t, err)

			got := out.String()
			tt.checkFunc(t, got)
		})
	}
}
