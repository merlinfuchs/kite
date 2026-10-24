package eval

import (
	"context"
	"strings"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := map[string]any{
		"10":                            10,
		" 5 + 5 ":                       10,
		"(3 * 4) - 2":                   10,
		"2x5":                           10,
		"20 ÷ 2":                        10,
		"2^3^2":                         512,
		"-2^2":                          -4,
		"--3":                           3,
		"7 % 4":                         3,
		"3/2":                           1.5,
		"0.1 * 30":                      3,
		"24/7":                          3.4285714285714284,
		"007":                           7,
		"1/0":                           nil,
		"9^9^9":                         nil,
		"5 + 5 lol":                     nil,
		"hello":                         nil,
		"":                              nil,
		"()":                            nil,
		"(1 + 2":                        nil,
		"1 2":                           nil,
		"1e5":                           nil,
		"1..5":                          nil,
		"١٢":                            nil,
		strings.Repeat("1+", 100) + "1": nil,
	}
	for text, want := range tests {
		if got := calculate(text); got != want {
			t.Errorf("calculate(%q) = %#v, want %#v", text, got, want)
		}
	}
}

func TestCalculateInExpression(t *testing.T) {
	c := NewContext(Env{"content": "4 + 6", "number": 7})
	got, err := Eval(context.Background(), "calculate(content) == 10 && calculate(number) == 7 && calculate('hi') == nil && calculate(missing) == nil", c)
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != true {
		t.Errorf("got %v", got.Value)
	}
}
