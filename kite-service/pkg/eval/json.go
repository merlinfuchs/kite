package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// ErrInvalidJSONTemplate is returned when a JSON template does not produce
// valid JSON once its placeholders have been evaluated.
var ErrInvalidJSONTemplate = errors.New("invalid JSON body")

// EvalJSONTemplate evaluates the placeholders of a JSON document and returns
// the resulting JSON.
//
// Unlike EvalTemplate it knows where in the document each placeholder sits:
//
//   - Inside a JSON string, the result is escaped, so a value containing
//     quotes, backslashes or newlines can no longer break the document.
//     {"content": "Hi {{user.name}}"} stays valid for any name.
//   - Outside a string, the result is inserted as a typed JSON value, so
//     numbers, booleans, arrays and objects keep their type.
//     {"count": {{result('x').status_code}}} produces a number, and a
//     placeholder that evaluates to nothing produces null.
//
// The template itself does not have to be valid JSON (an unquoted placeholder
// isn't), but the result does.
func EvalJSONTemplate(ctx context.Context, template string, c Context) ([]byte, error) {
	if len(template) > MaxTemplateLength {
		return nil, fmt.Errorf(
			"eval error: %w: %d characters, limit is %d",
			ErrTemplateTooLong, len(template), MaxTemplateLength,
		)
	}

	out, err := expandJSONTemplate(template, func(expression string) (thing.Thing, error) {
		return Eval(ctx, expression, c)
	})
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}

	var v any
	if err := json.Unmarshal(out, &v); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidJSONTemplate, err)
	}

	return out, nil
}

// expandJSONTemplate does the scanning for EvalJSONTemplate. It's split out so
// it can be tested without the expression engine.
func expandJSONTemplate(template string, evalFn func(expression string) (thing.Thing, error)) ([]byte, error) {
	var out bytes.Buffer
	out.Grow(len(template))

	// Only what the placeholders expand to has to be tracked, the literal
	// parts are already bounded by MaxTemplateLength.
	var emitted int

	inString := false
	escaped := false

	for i := 0; i < len(template); {
		if strings.HasPrefix(template[i:], templateStartTag) {
			rest := template[i+len(templateStartTag):]
			end := strings.Index(rest, templateEndTag)
			if end == -1 {
				return nil, fmt.Errorf("eval error: unclosed placeholder at offset %d", i)
			}

			res, err := evalFn(rest[:end])
			if err != nil {
				return nil, err
			}

			var val []byte
			if inString {
				val, err = jsonStringContent(templateString(res))
			} else {
				val, err = thingToJSON(res)
			}
			if err != nil {
				return nil, fmt.Errorf("eval error: failed to encode placeholder result: %w", err)
			}

			emitted += len(val)
			if emitted > MaxTemplateOutputLength {
				return nil, fmt.Errorf(
					"eval error: %w: expanded to at least %d bytes, limit is %d",
					ErrTemplateOutputTooLong, emitted, MaxTemplateOutputLength,
				)
			}

			out.Write(val)
			escaped = false
			i += len(templateStartTag) + end + len(templateEndTag)
			continue
		}

		ch := template[i]
		out.WriteByte(ch)

		if inString {
			switch {
			case escaped:
				escaped = false
			case ch == '\\':
				escaped = true
			case ch == '"':
				inString = false
			}
		} else if ch == '"' {
			inString = true
		}

		i++
	}

	return out.Bytes(), nil
}

// templateString is how a placeholder result is rendered inside a string,
// matching what EvalTemplate would produce for it, except that plain
// arrays and objects are rendered as JSON instead of Go's %v format.
func templateString(t thing.Thing) string {
	if t.IsNil() {
		return ""
	}

	switch t.Type {
	case thing.TypeArray, thing.TypeObject:
		if b, err := thingToJSON(t); err == nil {
			return string(b)
		}
	case thing.TypeAny:
		switch t.Value.(type) {
		case map[string]any, []any:
			if b, err := marshalJSON(t.Value); err == nil {
				return string(b)
			}
		}
	}

	return fmt.Sprintf("%v", t)
}

// jsonStringContent escapes s for use inside a JSON string, without the
// surrounding quotes.
func jsonStringContent(s string) ([]byte, error) {
	b, err := marshalJSON(s)
	if err != nil {
		return nil, err
	}
	return b[1 : len(b)-1], nil
}

// ThingToJSON converts a result into a standalone JSON value, the same way a
// placeholder outside quotes in a JSON body is converted.
func ThingToJSON(t thing.Thing) ([]byte, error) {
	return thingToJSON(t)
}

// thingToJSON converts a placeholder result into a standalone JSON value.
func thingToJSON(t thing.Thing) ([]byte, error) {
	if t.IsNil() {
		return []byte("null"), nil
	}

	switch t.Type {
	case thing.TypeString, thing.TypeInt, thing.TypeFloat, thing.TypeBool:
		return marshalJSON(t.Value)
	case thing.TypeHTTPResponse:
		// Pass a JSON response straight through, otherwise send it as text.
		body := bytes.TrimSpace(t.HTTPResponse().Body)
		if len(body) > 0 && json.Valid(body) {
			return body, nil
		}
		return marshalJSON(string(t.HTTPResponse().Body))
	case thing.TypeArray:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, v := range t.Array() {
			if i > 0 {
				buf.WriteByte(',')
			}
			b, err := thingToJSON(v)
			if err != nil {
				return nil, err
			}
			buf.Write(b)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	case thing.TypeObject:
		obj := make(map[string]json.RawMessage, len(t.Object()))
		for k, v := range t.Object() {
			b, err := thingToJSON(v)
			if err != nil {
				return nil, err
			}
			obj[k] = b
		}
		return marshalJSON(obj)
	default:
		b, err := marshalJSON(t.Value)
		if err != nil {
			// Not everything the expression engine returns can be encoded,
			// fall back to how it would print in a message.
			return marshalJSON(t.String())
		}
		return b, nil
	}
}

// marshalJSON is json.Marshal without escaping <, > and &, which are valid
// in JSON and only make the body harder to read on the receiving end.
func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
