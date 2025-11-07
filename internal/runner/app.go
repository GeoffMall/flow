package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GeoffMall/flow/internal/cli"
	"github.com/GeoffMall/flow/internal/format"
	_ "github.com/GeoffMall/flow/internal/format/avro"    // Register Avro format
	_ "github.com/GeoffMall/flow/internal/format/json"    // Register JSON format
	_ "github.com/GeoffMall/flow/internal/format/parquet" // Register Parquet format
	_ "github.com/GeoffMall/flow/internal/format/yaml"    // Register YAML format
	"github.com/GeoffMall/flow/internal/operation"
)

func Run() {
	f := cli.ParseFlags()
	// Enable color by default unless --no-color is specified
	f.Color = !f.NoColor

	// Handle directory mode
	if f.InputDir != "" {
		if err := processDirectory(f); err != nil {
			fatalf("Directory processing error: %v\n", err)
		}
		return
	}

	// Handle single file/stdin mode
	in, inClose, err := openInput(f.InputFile)
	if err != nil {
		fatalf("Error opening input: %v\n", err)
	}
	defer inClose()

	out, outClose, err := openOutput(f.OutputFile)
	if err != nil {
		fatalf("Error opening output: %v\n", err)
	}
	defer outClose()

	if err := run(in, out, f); err != nil {
		fatalf("Processing error: %v\n", err)
	}
}

func openInput(path string) (io.Reader, func(), error) {
	if path == "" {
		return os.Stdin, func() {}, nil
	}
	// #nosec G304 - CLI tool trusts user-provided file paths
	f, err := os.Open(path)
	if err != nil {
		return nil, func() {}, err
	}
	return f, func() { _ = f.Close() }, nil
}

func openOutput(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	// #nosec G304 - CLI tool trusts user-provided file paths
	f, err := os.Create(path)
	if err != nil {
		return nil, func() {}, err
	}
	return f, func() { _ = f.Close() }, nil
}

// processDirectory processes all files in a directory that match the specified format.
// It walks the directory tree, filters files by extension, and processes each matching file.
// Errors are collected and reported at the end (continue-on-error behavior).
//
//nolint:cyclop,funlen // Directory walking requires multiple error checks and comprehensive handling
func processDirectory(opts *cli.Flags) error {
	// Determine which extensions to process based on -from flag
	var extensions []string
	switch opts.FromFormat {
	case "avro":
		extensions = []string{".avro"}
	case "parquet":
		extensions = []string{".parquet"}
	case "yaml":
		extensions = []string{".yaml", ".yml"}
	case "json", "":
		extensions = []string{".json"}
	default:
		return fmt.Errorf("unknown format for directory processing: %s", opts.FromFormat)
	}

	// Open output once for all files
	out, outClose, err := openOutput(opts.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to open output: %w", err)
	}
	defer outClose()

	// Collect errors from processing
	var errors []error
	fileCount := 0

	// Walk the directory
	err = filepath.WalkDir(opts.InputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			errors = append(errors, fmt.Errorf("error accessing %s: %w", path, err))
			return nil // Continue processing other files
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file extension matches
		ext := strings.ToLower(filepath.Ext(path))
		matched := false
		for _, allowedExt := range extensions {
			if ext == allowedExt {
				matched = true
				break
			}
		}
		if !matched {
			return nil // Skip files that don't match
		}

		fileCount++

		// Process this file
		// #nosec G304 - CLI tool processes user-specified directory paths
		file, err := os.Open(path)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to open %s: %w", path, err))
			return nil // Continue processing other files
		}
		defer file.Close()

		// Process the file with metadata (filename and row tracking)
		if err := runWithMetadata(file, out, opts, path); err != nil {
			errors = append(errors, fmt.Errorf("failed to process %s: %w", path, err))
			return nil // Continue processing other files
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	if fileCount == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: no files with extensions %v found in %s\n", extensions, opts.InputDir)
	}

	// Report collected errors
	if len(errors) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "\nEncountered %d error(s) during processing:\n", len(errors))
		for i, e := range errors {
			_, _ = fmt.Fprintf(os.Stderr, "  %d. %v\n", i+1, e)
		}
		return fmt.Errorf("directory processing completed with %d error(s)", len(errors))
	}

	return nil
}

func buildPipeline(opts *cli.Flags) (*operation.Pipeline, error) {
	var ops []operation.Operation

	// WHERE filtering comes first to filter out non-matching documents early
	if len(opts.WherePairs) > 0 {
		whereOp, err := operation.NewWhere(opts.WherePairs)
		if err != nil {
			return nil, err
		}
		ops = append(ops, whereOp)
	}

	if len(opts.PickPaths) > 0 {
		ops = append(ops, operation.NewPick(opts.PickPaths, opts.PreserveHierarchy))
	}

	if len(opts.SetPairs) > 0 {
		setOp, err := operation.NewSetFromPairs(opts.SetPairs)
		if err != nil {
			return nil, err
		}
		ops = append(ops, setOp)
	}

	if len(opts.DeletePaths) > 0 {
		ops = append(ops, operation.NewDelete(opts.DeletePaths))
	}

	return operation.NewPipeline(ops...), nil
}

// determineInputFormat determines the input format based on flags and file extension.
// Priority: explicit -from flag > file extension > default (json)
func determineInputFormat(opts *cli.Flags) string {
	// If explicit format specified, use it
	if opts.FromFormat != "" {
		return opts.FromFormat
	}

	// If reading from a file, check extension
	if opts.InputFile != "" {
		ext := strings.ToLower(filepath.Ext(opts.InputFile))
		if ext == ".yaml" || ext == ".yml" {
			return "yaml"
		}
		if ext == ".avro" {
			return "avro"
		}
		if ext == ".parquet" {
			return "parquet"
		}
	}

	// Default to JSON
	return "json"
}

// run executes one full pass: parse stream -> apply pipeline -> print.
func run(in io.Reader, out io.Writer, opts *cli.Flags) error {
	// Build operation pipeline
	pipe, err := buildPipeline(opts)
	if err != nil {
		return err
	}

	// Determine input format
	inputFormatName := determineInputFormat(opts)

	// Get input format
	inputFormat, err := format.Get(inputFormatName)
	if err != nil {
		return fmt.Errorf("unknown input format %q: %w", inputFormatName, err)
	}

	// Create parser
	parser, err := inputFormat.NewParser(in)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	// Determine output format (default to json if not specified)
	outputFormatName := opts.ToFormat
	if outputFormatName == "" {
		outputFormatName = "json"
	}

	// Get output format
	outputFormat, err := format.Get(outputFormatName)
	if err != nil {
		return fmt.Errorf("unknown output format %q: %w", outputFormatName, err)
	}

	// Create formatter for output
	formatter := outputFormat.NewFormatter(out, format.FormatterOptions{
		Color:   opts.Color,
		Compact: opts.Compact,
	})
	defer formatter.Close()

	// Process stream: parse -> transform -> format
	return parser.ForEach(func(doc any) error {
		outDoc := doc
		if !pipe.Empty() {
			var err error
			outDoc, err = pipe.Apply(doc)
			if err != nil {
				return err
			}
		}
		// Skip if document was filtered out (e.g., by WHERE operation)
		if outDoc == operation.Filtered {
			return nil
		}
		return formatter.Write(outDoc)
	})
}

// runWithMetadata is like run but wraps each output document with metadata (_file, _row, data).
// This is used when processing directories to track which file and row each result came from.
func runWithMetadata(in io.Reader, out io.Writer, opts *cli.Flags, filename string) error {
	// Build operation pipeline
	pipe, err := buildPipeline(opts)
	if err != nil {
		return err
	}

	// Determine input format
	inputFormatName := determineInputFormat(opts)

	// Get input format
	inputFormat, err := format.Get(inputFormatName)
	if err != nil {
		return fmt.Errorf("unknown input format %q: %w", inputFormatName, err)
	}

	// Create parser
	parser, err := inputFormat.NewParser(in)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	// Determine output format (default to json if not specified)
	outputFormatName := opts.ToFormat
	if outputFormatName == "" {
		outputFormatName = "json"
	}

	// Get output format
	outputFormat, err := format.Get(outputFormatName)
	if err != nil {
		return fmt.Errorf("unknown output format %q: %w", outputFormatName, err)
	}

	// Create formatter for output
	formatter := outputFormat.NewFormatter(out, format.FormatterOptions{
		Color:   opts.Color,
		Compact: opts.Compact,
	})
	defer formatter.Close()

	// Track row number
	rowNum := 0

	// Process stream: parse -> transform -> wrap with metadata -> format
	return parser.ForEach(func(doc any) error {
		rowNum++

		outDoc := doc
		if !pipe.Empty() {
			var err error
			outDoc, err = pipe.Apply(doc)
			if err != nil {
				return err
			}
		}

		// Skip if document was filtered out (e.g., by WHERE operation)
		if outDoc == operation.Filtered {
			return nil
		}

		// Wrap with metadata
		wrapped := map[string]any{
			"_file": filename,
			"_row":  rowNum,
			"data":  outDoc,
		}

		return formatter.Write(wrapped)
	})
}

func fatalf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
