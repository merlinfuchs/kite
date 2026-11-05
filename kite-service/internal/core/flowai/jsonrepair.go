package flowai

import (
	"encoding/json"
	"log/slog"
	"strings"
)

// unmarshalModelJSON reads JSON the model wrote in a string field. If it
// isn't valid, the most common mistakes are repaired first, so they don't
// cost a repair round. The error is the one of the original value.
func unmarshalModelJSON(value string, v any) error {
	err := json.Unmarshal([]byte(value), v)
	if err == nil {
		return nil
	}
	repaired := repairJSON(value)
	if repaired == value || json.Unmarshal([]byte(repaired), v) != nil {
		return err
	}
	// Logged to see how often the model gets JSON wrong.
	slog.Info("Flow AI edit JSON repaired", slog.Any("cause", err))
	return nil
}

// repairJSON fixes the mistakes the model makes most when writing JSON:
// raw line breaks and tabs in strings, \' and other invalid escapes,
// closing brackets that don't match the open ones, and missing closing
// brackets at the end. Strings that aren't closed are left as they are.
func repairJSON(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)

	var open []byte
	inString := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if inString {
			switch c {
			case '"':
				inString = false
				b.WriteByte(c)
			case '\\':
				if i+1 >= len(s) {
					continue
				}
				next := s[i+1]
				switch next {
				case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
					b.WriteByte(c)
					b.WriteByte(next)
				case 'u':
					if i+5 < len(s) && isHex(s[i+2:i+6]) {
						b.WriteString(`\u`)
					} else {
						b.WriteString(`\\u`)
					}
				case '\'':
					b.WriteByte(next)
				default:
					// Keep the backslash as text.
					b.WriteString(`\\`)
					b.WriteByte(next)
				}
				i++
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\t':
				b.WriteString(`\t`)
			default:
				b.WriteByte(c)
			}
			continue
		}

		switch c {
		case '"':
			inString = true
		case '{', '[':
			open = append(open, c)
		case '}', ']':
			at := -1
			for j := len(open) - 1; j >= 0; j-- {
				if closer(open[j]) == c {
					at = j
					break
				}
			}
			if at == -1 {
				// Nothing to close, so the bracket is dropped.
				continue
			}
			// Close what was left open inside, e.g. the ] of
			// {"a": [{"b": 1}} before its }.
			for j := len(open) - 1; j > at; j-- {
				b.WriteByte(closer(open[j]))
			}
			open = open[:at]
		}
		b.WriteByte(c)
	}

	// Where a string that isn't closed should end can't be told, e.g. in
	// {"content":"Hello}}, so it's left for the model to fix.
	if inString {
		return s
	}
	for j := len(open) - 1; j >= 0; j-- {
		b.WriteByte(closer(open[j]))
	}

	return b.String()
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i] | 0x20
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func closer(opener byte) byte {
	if opener == '[' {
		return ']'
	}
	return '}'
}
