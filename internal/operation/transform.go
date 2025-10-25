package operation

import "fmt"

// A minimal pipeline that applies a sequence of operations to a document.
// Each Operation is expected to transform and return a new value (or the same one).
// The pipeline runs them in order and returns the final result.

// Pipeline runs multiple operations in sequence.
type Pipeline struct {
	Ops []Operation
}

// NewPipeline constructs a pipeline from a variadic list of ops.
func NewPipeline(ops ...Operation) *Pipeline {
	// Defensive copy so callers can reuse/slice their own list safely.
	cpy := make([]Operation, len(ops))
	copy(cpy, ops)
	return &Pipeline{Ops: cpy}
}

// Append adds more operations to the pipeline.
func (p *Pipeline) Append(ops ...Operation) {
	p.Ops = append(p.Ops, ops...)
}

// Empty reports whether the pipeline has any ops.
func (p *Pipeline) Empty() bool { return len(p.Ops) == 0 }

// Apply runs every operation in order, passing the output of each as the
// input to the next. If any step fails, it returns a StepError describing
// which operation failed and why.
func (p *Pipeline) Apply(v any) (any, error) {
	current := v
	for i, op := range p.Ops {
		next, err := op.Apply(current)
		if err != nil {
			return nil, StepError{
				Index:   i,
				OpDesc:  safeDesc(op),
				Wrapped: err,
			}
		}
		current = next
	}
	return current, nil
}

// Compose is a convenience that constructs a pipeline and applies it immediately.
func Compose(v any, ops ...Operation) (any, error) {
	return NewPipeline(ops...).Apply(v)
}

// --------------------------- Error types ---------------------------

// StepError annotates an error with pipeline position and op description.
type StepError struct {
	Index   int
	OpDesc  string
	Wrapped error
}

func (e StepError) Error() string {
	if e.OpDesc == "" {
		return fmt.Sprintf("pipeline step %d failed: %v", e.Index, e.Wrapped)
	}
	return fmt.Sprintf("pipeline step %d (%s) failed: %v", e.Index, e.OpDesc, e.Wrapped)
}

func (e StepError) Unwrap() error { return e.Wrapped }

// safeDesc guards against panics in Description() (defensive; unlikely).
func safeDesc(op Operation) (desc string) {
	defer func() {
		if recover() != nil {
			desc = ""
		}
	}()
	return op.Description()
}
