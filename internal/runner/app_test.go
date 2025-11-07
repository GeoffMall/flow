package runner

import (
	"bytes"
	"os"
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

func Test_processDirectory_Avro_MultipleFiles(t *testing.T) {
	opts := &cli.Flags{
		InputDir:   "../../testdata/dir-test",
		FromFormat: "avro",
		Compact:    true,
		NoColor:    true,
	}

	err := processDirectory(opts)
	assert.NoError(t, err)

	// Should process employees1.avro (5 records) + employees2.avro (5 records) + users.avro (3 records) = 13 total
	// Note: We can't easily capture output from processDirectory since it uses openOutput
	// This test mainly verifies no errors occur
}

func Test_processDirectory_Parquet_MultipleFiles(t *testing.T) {
	opts := &cli.Flags{
		InputDir:   "../../testdata/dir-test",
		FromFormat: "parquet",
		Compact:    true,
		NoColor:    true,
	}

	err := processDirectory(opts)
	assert.NoError(t, err)

	// Should process products1.parquet (5 records) + products2.parquet (5 records) + single.parquet (1 record) = 11 total
}

func Test_processDirectory_WithWhere_MultipleMatches(t *testing.T) {
	// This test verifies that WHERE filtering returns multiple matching records across multiple files
	// Using a temporary output file to capture results
	tmpFile := "../../testdata/dir-test/test-output.json"
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	opts := &cli.Flags{
		InputDir:   "../../testdata/dir-test",
		FromFormat: "avro",
		WherePairs: []string{"department=Engineering"},
		Compact:    true,
		NoColor:    true,
		OutputFile: tmpFile,
	}

	err := processDirectory(opts)
	assert.NoError(t, err)

	// Read the output file and verify multiple matches
	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	output := string(content)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 5 Engineering employees total:
	// - Alice Johnson (employees1.avro, row 1)
	// - Bob Smith (employees1.avro, row 2)
	// - Eve Davis (employees1.avro, row 5)
	// - Grace Lee (employees2.avro, row 2)
	// - Jack Anderson (employees2.avro, row 5)
	assert.Equal(t, 5, len(lines), "should return 5 Engineering employees")

	// Verify all lines contain department=Engineering
	for _, line := range lines {
		assert.Contains(t, line, `"department":"Engineering"`)
		assert.Contains(t, line, `"_file"`)
		assert.Contains(t, line, `"_row"`)
		assert.Contains(t, line, `"data"`)
	}

	// Verify we have results from both files
	assert.Contains(t, output, "employees1.avro")
	assert.Contains(t, output, "employees2.avro")
}

func Test_processDirectory_Parquet_WithWhere_MultipleMatches(t *testing.T) {
	tmpFile := "../../testdata/dir-test/test-parquet-output.json"
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	opts := &cli.Flags{
		InputDir:   "../../testdata/dir-test",
		FromFormat: "parquet",
		WherePairs: []string{"category=Electronics"},
		Compact:    true,
		NoColor:    true,
		OutputFile: tmpFile,
	}

	err := processDirectory(opts)
	assert.NoError(t, err)

	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	output := string(content)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 7 Electronics products total across products1.parquet and products2.parquet
	assert.GreaterOrEqual(t, len(lines), 7, "should return at least 7 Electronics products")

	// Verify all lines contain category=Electronics
	for _, line := range lines {
		assert.Contains(t, line, `"category":"Electronics"`)
	}

	// Verify we have results from both files
	assert.Contains(t, output, "products1.parquet")
	assert.Contains(t, output, "products2.parquet")
}

func Test_processDirectory_WithMultipleWhereConditions(t *testing.T) {
	tmpFile := "../../testdata/dir-test/test-multi-where.json"
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	opts := &cli.Flags{
		InputDir:   "../../testdata/dir-test",
		FromFormat: "avro",
		WherePairs: []string{"department=Engineering", "active=true"},
		Compact:    true,
		NoColor:    true,
		OutputFile: tmpFile,
	}

	err := processDirectory(opts)
	assert.NoError(t, err)

	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	output := string(content)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 5 active Engineering employees
	assert.Equal(t, 5, len(lines), "should return 5 active Engineering employees")

	// Verify all lines match both conditions
	for _, line := range lines {
		assert.Contains(t, line, `"department":"Engineering"`)
		assert.Contains(t, line, `"active":true`)
	}
}

func Test_processDirectory_NoMatchingFiles(t *testing.T) {
	opts := &cli.Flags{
		InputDir:   "../../testdata",
		FromFormat: "avro",
		Compact:    true,
		NoColor:    true,
	}

	err := processDirectory(opts)
	// Should succeed but warn about no matching files (warning goes to stderr)
	assert.NoError(t, err)
}

func Test_processDirectory_InvalidDirectory(t *testing.T) {
	opts := &cli.Flags{
		InputDir:   "../../testdata/nonexistent",
		FromFormat: "avro",
		Compact:    true,
		NoColor:    true,
	}

	err := processDirectory(opts)
	assert.Error(t, err)
}

func Test_processDirectory_InvalidFormat(t *testing.T) {
	opts := &cli.Flags{
		InputDir:   "../../testdata/dir-test",
		FromFormat: "unknown-format",
		Compact:    true,
		NoColor:    true,
	}

	err := processDirectory(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format")
}

func Test_processDirectory_OutputFileError(t *testing.T) {
	opts := &cli.Flags{
		InputDir:   "../../testdata/dir-test",
		FromFormat: "avro",
		OutputFile: "/nonexistent/directory/output.json",
		Compact:    true,
		NoColor:    true,
	}

	err := processDirectory(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open output")
}

func Test_run_YAMLToYAML(t *testing.T) {
	in := strings.NewReader(`name: Alice
age: 30`)
	var out bytes.Buffer

	opts := &cli.Flags{
		FromFormat: "yaml",
		ToFormat:   "yaml",
	}

	err := run(in, &out, opts)
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "name: Alice")
	assert.Contains(t, output, "age: 30")
}

func Test_run_AvroFormat(t *testing.T) {
	file, err := os.Open("../../testdata/dir-test/employees1.avro")
	assert.NoError(t, err)
	defer file.Close()

	var out bytes.Buffer
	opts := &cli.Flags{
		FromFormat: "avro",
		Compact:    true,
		NoColor:    true,
	}

	err = run(file, &out, opts)
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Alice Johnson")
	assert.Contains(t, output, "Engineering")
}

func Test_run_ParquetFormat(t *testing.T) {
	file, err := os.Open("../../testdata/dir-test/products1.parquet")
	assert.NoError(t, err)
	defer file.Close()

	var out bytes.Buffer
	opts := &cli.Flags{
		FromFormat: "parquet",
		Compact:    true,
		NoColor:    true,
	}

	err = run(file, &out, opts)
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Dell XPS 15")
	assert.Contains(t, output, "Electronics")
}

func Test_runWithMetadata_AvroFormat(t *testing.T) {
	file, err := os.Open("../../testdata/dir-test/employees1.avro")
	assert.NoError(t, err)
	defer file.Close()

	var out bytes.Buffer
	opts := &cli.Flags{
		FromFormat: "avro",
		Compact:    true,
		NoColor:    true,
	}

	err = runWithMetadata(file, &out, opts, "employees1.avro")
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"_file":"employees1.avro"`)
	assert.Contains(t, output, `"_row":`)
	assert.Contains(t, output, "Alice Johnson")
}

func Test_runWithMetadata_ParquetFormat(t *testing.T) {
	file, err := os.Open("../../testdata/dir-test/products1.parquet")
	assert.NoError(t, err)
	defer file.Close()

	var out bytes.Buffer
	opts := &cli.Flags{
		FromFormat: "parquet",
		Compact:    true,
		NoColor:    true,
	}

	err = runWithMetadata(file, &out, opts, "products1.parquet")
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"_file":"products1.parquet"`)
	assert.Contains(t, output, `"_row":`)
	assert.Contains(t, output, "Dell XPS 15")
}

//nolint:funlen // Table-driven test covering comprehensive JSON data types
func Test_determineInputFormat(t *testing.T) {
	tests := []struct {
		name        string
		opts        *cli.Flags
		expectedFmt string
	}{
		{
			name:        "explicit_json_flag",
			opts:        &cli.Flags{FromFormat: "json"},
			expectedFmt: "json",
		},
		{
			name:        "explicit_yaml_flag",
			opts:        &cli.Flags{FromFormat: "yaml"},
			expectedFmt: "yaml",
		},
		{
			name:        "explicit_avro_flag",
			opts:        &cli.Flags{FromFormat: "avro"},
			expectedFmt: "avro",
		},
		{
			name:        "explicit_parquet_flag",
			opts:        &cli.Flags{FromFormat: "parquet"},
			expectedFmt: "parquet",
		},
		{
			name:        "yaml_extension",
			opts:        &cli.Flags{InputFile: "config.yaml"},
			expectedFmt: "yaml",
		},
		{
			name:        "yml_extension",
			opts:        &cli.Flags{InputFile: "config.yml"},
			expectedFmt: "yaml",
		},
		{
			name:        "avro_extension",
			opts:        &cli.Flags{InputFile: "data.avro"},
			expectedFmt: "avro",
		},
		{
			name:        "parquet_extension",
			opts:        &cli.Flags{InputFile: "data.parquet"},
			expectedFmt: "parquet",
		},
		{
			name:        "json_extension",
			opts:        &cli.Flags{InputFile: "data.json"},
			expectedFmt: "json",
		},
		{
			name:        "no_extension_defaults_to_json",
			opts:        &cli.Flags{InputFile: "data"},
			expectedFmt: "json",
		},
		{
			name:        "no_input_file_defaults_to_json",
			opts:        &cli.Flags{},
			expectedFmt: "json",
		},
		{
			name:        "uppercase_yaml_extension",
			opts:        &cli.Flags{InputFile: "CONFIG.YAML"},
			expectedFmt: "yaml",
		},
		{
			name:        "explicit_flag_overrides_extension",
			opts:        &cli.Flags{InputFile: "data.json", FromFormat: "yaml"},
			expectedFmt: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineInputFormat(tt.opts)
			assert.Equal(t, tt.expectedFmt, got)
		})
	}
}

func Test_openInput_Error(t *testing.T) {
	_, _, err := openInput("/nonexistent/path/to/file.json")
	assert.Error(t, err)
}

func Test_openInput_Stdin(t *testing.T) {
	reader, closeFunc, err := openInput("")
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	assert.NotNil(t, closeFunc)
	closeFunc() // Should not panic
}

func Test_openOutput_Error(t *testing.T) {
	_, _, err := openOutput("/nonexistent/directory/output.json")
	assert.Error(t, err)
}

func Test_openOutput_Stdout(t *testing.T) {
	writer, closeFunc, err := openOutput("")
	assert.NoError(t, err)
	assert.NotNil(t, writer)
	assert.NotNil(t, closeFunc)
	closeFunc() // Should not panic
}

func Test_buildPipeline_WithWhere(t *testing.T) {
	opts := &cli.Flags{
		WherePairs: []string{"name=Alice", "age=30"},
	}
	pipe, err := buildPipeline(opts)
	assert.NoError(t, err)
	assert.NotNil(t, pipe)
	assert.False(t, pipe.Empty())
}

func Test_buildPipeline_Empty(t *testing.T) {
	opts := &cli.Flags{}
	pipe, err := buildPipeline(opts)
	assert.NoError(t, err)
	assert.NotNil(t, pipe)
	assert.True(t, pipe.Empty())
}

func Test_buildPipeline_InvalidWhere(t *testing.T) {
	opts := &cli.Flags{
		WherePairs: []string{"invalid-no-equals"},
	}
	_, err := buildPipeline(opts)
	assert.Error(t, err)
}

func Test_run_UnknownInputFormat(t *testing.T) {
	in := strings.NewReader(`{"test": "data"}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		FromFormat: "unknown-format",
		Compact:    true,
	}

	err := run(in, &out, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown input format")
}

func Test_run_UnknownOutputFormat(t *testing.T) {
	in := strings.NewReader(`{"test": "data"}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		ToFormat: "unknown-format",
		Compact:  true,
	}

	err := run(in, &out, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output format")
}

func Test_run_ParserCreationError(t *testing.T) {
	in := strings.NewReader(`not valid json at all {{{`)
	var out bytes.Buffer

	opts := &cli.Flags{
		Compact: true,
	}

	err := run(in, &out, opts)
	assert.Error(t, err)
}

func Test_run_WithWhere_FiltersOut(t *testing.T) {
	in := strings.NewReader(`{"name":"Alice","age":30}
{"name":"Bob","age":25}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		WherePairs: []string{"name=Alice"},
		Compact:    true,
		NoColor:    true,
	}

	err := run(in, &out, opts)
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Alice")
	assert.NotContains(t, output, "Bob")
}

func Test_runWithMetadata_Basic(t *testing.T) {
	in := strings.NewReader(`{"name":"Alice","age":30}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		Compact: true,
		NoColor: true,
	}

	err := runWithMetadata(in, &out, opts, "test.json")
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"_file":"test.json"`)
	assert.Contains(t, output, `"_row":1`)
	assert.Contains(t, output, `"data"`)
	assert.Contains(t, output, `"name":"Alice"`)
}

func Test_runWithMetadata_WithWhere(t *testing.T) {
	in := strings.NewReader(`{"name":"Alice","age":30}
{"name":"Bob","age":25}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		WherePairs: []string{"name=Bob"},
		Compact:    true,
		NoColor:    true,
	}

	err := runWithMetadata(in, &out, opts, "users.json")
	assert.NoError(t, err)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines), "should only return Bob")
	assert.Contains(t, output, `"name":"Bob"`)
	assert.Contains(t, output, `"_row":2`) // Bob is row 2 (even though Alice was filtered)
}

func Test_runWithMetadata_UnknownFormat(t *testing.T) {
	in := strings.NewReader(`{"test": "data"}`)
	var out bytes.Buffer

	opts := &cli.Flags{
		FromFormat: "unknown-format",
		Compact:    true,
	}

	err := runWithMetadata(in, &out, opts, "test.dat")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown input format")
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
