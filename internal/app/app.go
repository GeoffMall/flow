package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GeoffMall/flow/internal/cli"
	"github.com/GeoffMall/flow/internal/format"
	_ "github.com/GeoffMall/flow/internal/format/json" // Register JSON format
	_ "github.com/GeoffMall/flow/internal/format/yaml" // Register YAML format
	"github.com/GeoffMall/flow/internal/operation"
)

func Run() {
	f := cli.ParseFlags()
	// Enable color by default unless --no-color is specified
	f.Color = !f.NoColor

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

func buildPipeline(opts *cli.Flags) (*operation.Pipeline, error) {
	var ops []operation.Operation

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
		return formatter.Write(outDoc)
	})
}

func fatalf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
