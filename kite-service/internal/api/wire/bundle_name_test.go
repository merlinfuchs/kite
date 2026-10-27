package wire

import "testing"

func TestValidateBundleName(t *testing.T) {
	for name, ok := range map[string]bool{
		"Purge Messages":           true,
		"Moderation | Honeypot":    true,
		"Café & Co. (v2)":          true,
		"🗑️ | Purge":               false,
		"Guess the Number ✨":       false,
		"⛔ DNI Lists":              false,
		"Role-Dropping System w/ ": true,
	} {
		if err := validateBundleName(name); (err == nil) != ok {
			t.Errorf("validateBundleName(%q) = %v, want ok %v", name, err, ok)
		}
	}
}
