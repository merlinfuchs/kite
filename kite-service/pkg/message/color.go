package message

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Color is a color as Discord sends it, a number like 0x5865F2. It can also be
// read from text like "#5865F2", which the flow AI and other tools write.
type Color int

func (c *Color) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*c = Color(n)
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("color must be a number or a hex color like #5865F2")
	}

	s = strings.TrimSpace(s)
	if s == "" {
		*c = 0
		return nil
	}

	var (
		v   int64
		err error
	)
	if hex, ok := strings.CutPrefix(s, "#"); ok {
		v, err = strconv.ParseInt(hex, 16, 64)
	} else if hex, ok := strings.CutPrefix(strings.ToLower(s), "0x"); ok {
		v, err = strconv.ParseInt(hex, 16, 64)
	} else if v, err = strconv.ParseInt(s, 10, 64); err != nil && len(s) == 6 {
		// Hex without the #, like "ff0000".
		v, err = strconv.ParseInt(s, 16, 64)
	}
	if err != nil {
		return fmt.Errorf("color %q must be a number or a hex color like #5865F2", s)
	}

	*c = Color(v)
	return nil
}
