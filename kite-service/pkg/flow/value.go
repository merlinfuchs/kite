package flow

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Value is a wrapper around any with some helper methods
type Value struct {
	Inner any
}

func NewValue(v any) Value {
	return Value{Inner: v}
}

func (w Value) String() string {
	return fmt.Sprintf("%v", w.Inner)
}

func (w Value) Int() int64 {
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

func (w Value) Float() float64 {
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

func (w Value) Bool() bool {
	switch v := w.Inner.(type) {
	case bool:
		return v
	default:
		return w.Int() != 0
	}
}

func (w Value) Equals(other *Value) bool {
	return w.String() == other.String()
}

func (w Value) EqualsStrict(other *Value) bool {
	return reflect.DeepEqual(w.Inner, other.Inner)
}

func (w Value) GreaterThan(other *Value) bool {
	return w.Float() > other.Float()
}

func (w Value) GreaterThanOrEqual(other *Value) bool {
	return w.Float() >= other.Float()
}

func (w Value) LessThan(other *Value) bool {
	return w.Float() < other.Float()
}

func (w Value) LessThanOrEqual(other *Value) bool {
	return w.Float() <= other.Float()
}

func (w Value) Contains(other *Value) bool {
	// TODO: handle arrays and objects?
	return strings.Contains(w.String(), other.String())
}

func (w Value) IsEmpty() bool {
	return w.String() == ""
}
