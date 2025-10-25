package printer

// ------------------------- JSON Colorizer -------------------------
// A lightweight JSON colorizer that works on already-encoded JSON bytes.
// It uses a small state machine, coloring:
//   - object keys (in strings before a ':')
//   - string values
//   - numbers
//   - true/false/null
//   - punctuation ({}[],:)
// If the JSON is not valid, it simply attempts best-effort highlighting.

const (
	colReset   = "\x1b[0m"
	colKey     = "\x1b[38;5;33m"  // blue-ish for keys
	colStr     = "\x1b[38;5;34m"  // green for strings
	colNum     = "\x1b[38;5;214m" // orange for numbers
	colBoolNil = "\x1b[38;5;135m" // purple for true/false/null
	colPunct   = "\x1b[38;5;240m" // gray for punctuation
)

func colorizeJSON(in []byte) []byte {
	out := make([]byte, 0, len(in)+len(in)/4) // small headroom

	type ctxType int
	const (
		ctxRoot ctxType = iota
		ctxObj
		ctxArr
	)

	// Stack to determine if we are inside an object and whether the next string is a key.
	type objState struct {
		expectKey bool
	}
	var stack []objState
	push := func(s objState) { stack = append(stack, s) }
	pop := func() {
		if len(stack) > 0 {
			stack = stack[:len(stack)-1]
		}
	}
	top := func() *objState {
		if len(stack) == 0 {
			return nil
		}
		return &stack[len(stack)-1]
	}

	// Track whether inside a string and escaping.
	inStr := false
	esc := false

	// Helper to write colored rune/bytes
	write := func(s string) { out = append(out, s...) }
	writeByte := func(b byte) { out = append(out, b) }

	for i := 0; i < len(in); i++ {
		b := in[i]

		if inStr {
			// Inside string
			writeByte(b)
			if esc {
				esc = false
				continue
			}
			if b == '\\' {
				esc = true
				continue
			}
			if b == '"' {
				// end string
				write(colReset)

				// If we're in an object and we just wrote a key (before ':'), set expectKey=false
				if st := top(); st != nil && st.expectKey {
					// The next significant non-space should be ':'
				}
				inStr = false
			}
			continue
		}

		switch b {
		case '{':
			write(colPunct)
			writeByte(b)
			write(colReset)
			// entering object: next string we see is a key
			push(objState{expectKey: true})
		case '}':
			write(colPunct)
			writeByte(b)
			write(colReset)
			// leaving object
			pop()
			// After a '}', if we are in object, next thing could be either ',' or end-of-object; if another key, expectKey=true will be set after ','
		case '[':
			write(colPunct)
			writeByte(b)
			write(colReset)
			// entering array doesn't affect expectKey
			push(objState{expectKey: false})
		case ']':
			write(colPunct)
			writeByte(b)
			write(colReset)
			pop()
		case ':', ',':
			write(colPunct)
			writeByte(b)
			write(colReset)
			if b == ',' {
				// After a comma inside an object, expect a key again.
				if st := top(); st != nil {
					st.expectKey = (len(stack) > 0 && st != nil && st.expectKey) // keep current
					// Actually, in object context, after ',', we expect a key; in array, nothing special.
					// We can't easily tell if we're in object or array from objState alone, but:
					// heuristic: if top exists and previously expectKey might have been false after a value, reset to true.
					st.expectKey = true
				}
			}
		case '"':
			// String start: color based on context (key vs value)
			if st := top(); st != nil && st.expectKey {
				write(colKey)
			} else {
				write(colStr)
			}
			writeByte(b)
			inStr = true

			// If we colored as key, we keep expectKey=true until we see ':'.
			// We'll toggle expectKey=false when ':' is encountered.
		case 't':
			// true
			if tryWord(in, &i, "true", &out, colBoolNil) {
				continue
			}
			writeByte(b)
		case 'f':
			// false
			if tryWord(in, &i, "false", &out, colBoolNil) {
				continue
			}
			writeByte(b)
		case 'n':
			// null
			if tryWord(in, &i, "null", &out, colBoolNil) {
				continue
			}
			writeByte(b)
		default:
			// numbers / spaces / others
			if isDigitOrNumberChar(b) {
				// color continuous number run
				write(colNum)
				j := i
				for j < len(in) && isDigitOrNumberChar(in[j]) {
					j++
				}
				out = append(out, in[i:j]...)
				write(colReset)
				i = j - 1
			} else {
				// space or other punctuation
				writeByte(b)
			}
		}

		// After writing ':', set expectKey=false (value next) for object context.
		if b == ':' {
			if st := top(); st != nil {
				st.expectKey = false
			}
		}
		// After ',', if we're in an object, expect a key next.
		if b == ',' {
			if st := top(); st != nil {
				st.expectKey = true
			}
		}
	}

	// Ensure newline (some compact JSON may not have one)
	if len(out) == 0 || out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return out
}

func isDigitOrNumberChar(b byte) bool {
	switch b {
	case '-', '+', '.', 'e', 'E':
		return true
	default:
		return b >= '0' && b <= '9'
	}
}

func tryWord(in []byte, i *int, word string, out *[]byte, color string) bool {
	if hasWordAt(in, *i, word) {
		*out = append(*out, color...)
		*out = append(*out, in[*i:*i+len(word)]...)
		*out = append(*out, colReset...)
		*i += len(word) - 1
		return true
	}
	return false
}

func hasWordAt(b []byte, i int, w string) bool {
	if i+len(w) > len(b) {
		return false
	}
	for j := 0; j < len(w); j++ {
		if b[i+j] != w[j] {
			return false
		}
	}
	return true
}
