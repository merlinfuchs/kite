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
type Color int

func (c *Color) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] != '"' {
		// Numbers aren't checked against maxColor, so stored data that
		// Discord would reject still loads, like the int this used to be.
		var f float64
		if err := json.Unmarshal(b, &f); err != nil || f != math.Trunc(f) || math.Abs(f) >= math.MaxInt64 {
			return fmt.Errorf("color must be a number or a hex color like #5865F2")
		}
		*c = Color(f)
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	v, err := parseColor(strings.TrimSpace(s))
	if err != nil {
		return err
	}
	*c = Color(v)
	return nil
}

// parseColor reads "#5865F2", "#fff", "0x5865F2" or "5793266".
func parseColor(s string) (int64, error) {
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
