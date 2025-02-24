package thing

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var Null = New(nil)

// Any is a wrapper around any with some helper methods
type Any struct {
	Inner any
}

func New(v any) Any {
	if a, ok := v.(Any); ok {
		return a
	}
	return Any{Inner: v}
}

func (w *Any) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &w.Inner)
}

func (w Any) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Inner)
}

func (w Any) String() string {
	return fmt.Sprintf("%v", w.Inner)
}

func (w Any) Int() int64 {
	switch v := w.Inner.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		i, err := strconv.ParseInt(w.String(), 10, 64)
		if err != nil {
			return 0
		}
		return i
	}
}

func (w Any) Float() float64 {
	switch v := w.Inner.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		i, err := strconv.ParseFloat(w.String(), 64)
		if err != nil {
			return 0
		}
		return i
	}
}

func (w Any) Bool() bool {
	switch v := w.Inner.(type) {
	case bool:
		return v
	default:
		return w.Int() != 0
	}
}

func (w Any) Equals(other *Any) bool {
	return w.String() == other.String()
}

func (w Any) EqualsStrict(other *Any) bool {
	return reflect.DeepEqual(w.Inner, other.Inner)
}

func (w Any) GreaterThan(other *Any) bool {
	return w.Float() > other.Float()
}

func (w Any) GreaterThanOrEqual(other *Any) bool {
	return w.Float() >= other.Float()
}

func (w Any) LessThan(other *Any) bool {
	return w.Float() < other.Float()
}

func (w Any) LessThanOrEqual(other *Any) bool {
	return w.Float() <= other.Float()
}

func (w Any) Contains(other *Any) bool {
	// TODO: handle arrays and objects?
	return strings.Contains(w.String(), other.String())
}

func (w Any) IsEmpty() bool {
	return w.String() == ""
}

func (w Any) IsNil() bool {
	return w.Inner == nil
}

func (w Any) Append(other Any) Any {
	// TODO: implement for arrays
	return New(w.String() + other.String())
}

func (w Any) Add(other Any) Any {
	return New(w.Float() + other.Float())
}

func (w Any) Sub(other Any) Any {
	return New(w.Float() - other.Float())
}

func Cast[T any](v Any) (T, bool) {
	if t, ok := v.Inner.(T); ok {
		return t, true
	}
	return *new(T), false
}
