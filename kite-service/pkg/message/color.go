package message

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// maxColor is the largest color Discord accepts, white.
const maxColor = 0xFFFFFF

// Color is a color as Discord sends it, a number like 0x5865F2. It can also be
// read from text like "#5865F2", which the flow AI and other tools write.
type Color struct {
	Value int
	// Template is a color with placeholders, like {{arg('color')}}, and Value
	// is 0 until a flow fills it in. Messages sent without one have no color.
	Template string
}

func (c Color) MarshalJSON() ([]byte, error) {
	if c.Template != "" {
		return json.Marshal(c.Template)
	}
	return json.Marshal(c.Value)
}

func (c *Color) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] != '"' {
		// Numbers aren't checked against maxColor, so stored data that
		// Discord would reject still loads, like the int this used to be.
		var f float64
		if err := json.Unmarshal(b, &f); err != nil || f != math.Trunc(f) || math.Abs(f) >= math.MaxInt64 {
			return fmt.Errorf("color must be a number or a hex color like #5865F2")
		}
		*c = Color{Value: int(f)}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	s = strings.TrimSpace(s)
	if strings.Contains(s, "{{") {
		*c = Color{Template: s}
		return nil
	}

	v, err := ParseColor(s)
	if err != nil {
		return err
	}
	*c = Color{Value: int(v)}
	return nil
}

// fill sets the color to what its template evaluates to, and returns false if
// that's empty.
func (c *Color) fill(eval func(template string) (string, error)) (bool, error) {
	if c.Template == "" {
		return true, nil
	}

	filled, err := eval(c.Template)
	if err != nil {
		return false, err
	}
	filled = strings.TrimSpace(filled)
	if filled == "" {
		*c = Color{}
		return false, nil
	}

	v, err := ParseColor(filled)
	if err != nil {
		return false, err
	}
	*c = Color{Value: int(v)}
	return true, nil
}

// ParseColor reads "#5865F2", "#fff", "0x5865F2" or "5793266".
func ParseColor(s string) (int64, error) {
	invalid := fmt.Errorf("color %q must be a number or a hex color like #5865F2", s)

	hex, ok := strings.CutPrefix(s, "#")
	if !ok {
		hex, ok = strings.CutPrefix(strings.ToLower(s), "0x")
	}
	if !ok {
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil || v > maxColor {
			return 0, invalid
		}
		return int64(v), nil
	}

	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, invalid
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, invalid
	}
	return int64(v), nil
}
