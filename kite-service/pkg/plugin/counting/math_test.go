package counting

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestParseCount(t *testing.T) {
	tests := []struct {
		content    string
		want       int64
		expression bool
	}{
		{"10", 10, false},
		{" 7 ", 7, false},
		{"5+5", 10, true},
		{"4+5", 9, true},
		{"(2+3)*4", 20, true},
		{"2^10", 1024, true},
		{"2**3", 8, true},
		{"100/4", 25, true},
		{"17 % 5", 2, true},
		{"sqrt(144)", 12, true},
		{"pow(2, 5) - 1", 31, true},
		{"SQRT(81) + 1", 10, true},
		{"3 × 4", 12, true},
		{"20 ÷ 4", 5, true},
		{"0.1*30", 3, true},
		{"round(pi * 10)", 31, true},
		{"floor(e * 4)", 10, true},
		{"abs(-3) + max(1, 6)", 9, true},
		{"log10(1000)", 3, true},
	}

	for _, tt := range tests {
		got, isExpr, err := parseCount(tt.content)
		if err != nil {
			t.Errorf("parseCount(%q): %v", tt.content, err)
			continue
		}
		if isExpr != tt.expression {
			t.Errorf("parseCount(%q) expression = %v, want %v", tt.content, isExpr, tt.expression)
		}
		if !matchesCount(got, tt.want) {
			t.Errorf("parseCount(%q) = %v, want %d", tt.content, got, tt.want)
		}
	}
}

func TestParseCountIgnoresChat(t *testing.T) {
	ignored := []string{
		"",
		"hello",
		"nice one",
		"lol 5",
		"5 apples",
		"pi",
		"1..1000000000",
		`len("aaaa")`,
		"1 == 1",
		"true ? 1 : 2",
		"upper(1)",
		"foo(1)",
		"5 % 0",
		strings.Repeat("1+", 150) + "1",
	}

	for _, content := range ignored {
		_, _, err := parseCount(content)
		if !errors.Is(err, errNotMath) {
			t.Errorf("parseCount(%q) err = %v, want errNotMath", content, err)
		}
	}
}

func TestMatchesCount(t *testing.T) {
	if !matchesCount(10, 10) {
		t.Error("10 should match 10")
	}
	if matchesCount(9, 10) {
		t.Error("9 should not match 10")
	}
	if matchesCount(2.5, 2) {
		t.Error("2.5 should not match 2")
	}
	if matchesCount(math.Inf(1), 1) {
		t.Error("inf should not match")
	}
}
