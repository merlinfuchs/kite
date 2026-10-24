package eval

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/expr-lang/expr"
)

// maxCalculateLength bounds the text calculate parses. Arithmetic people
// write in a message is far shorter, and it keeps the recursion shallow.
const maxCalculateLength = 200

// calculateFunc is calculate(text) in expressions.
var calculateFunc = expr.Function("calculate", func(params ...any) (any, error) {
	// Undefined variables arrive as nil.
	if params[0] == nil {
		return nil, nil
	}
	return calculate(fmt.Sprint(params[0])), nil
}, new(func(any) any))

// calculate returns the value of arithmetic like "(3 * 4) - 2", or nil if the
// text is anything else, like a chat message. The text comes from users, like
// the messages of a counting channel, so it's parsed here instead of by expr,
// which would also run functions, ranges and comments, and read hex and
// exponent literals. It knows + - * / % ^, x and × for times, ÷, and
// parentheses. Whole results are ints, so they print without a decimal point,
// also when floats are a little off, like 0.1 * 30.
func calculate(text string) any {
	if len(text) > maxCalculateLength {
		return nil
	}

	p := calcParser{text: []rune(text)}
	v := p.sum()
	p.skipSpace()
	if p.bad || p.pos != len(p.text) || math.IsInf(v, 0) || math.IsNaN(v) {
		return nil
	}

	if r := math.Round(v); math.Abs(v-r) < 1e-9 && math.Abs(r) < 1<<53 {
		return int(r)
	}
	return v
}

// calcParser stops at the first problem by setting bad, the values it returns
// after that don't matter.
type calcParser struct {
	text []rune
	pos  int
	bad  bool
}

func (p *calcParser) skipSpace() {
	for p.pos < len(p.text) && unicode.IsSpace(p.text[p.pos]) {
		p.pos++
	}
}

// next consumes the next rune if it's one of ops, and returns 0 otherwise.
func (p *calcParser) next(ops string) rune {
	p.skipSpace()
	if p.bad || p.pos == len(p.text) || !strings.ContainsRune(ops, p.text[p.pos]) {
		return 0
	}
	p.pos++
	return p.text[p.pos-1]
}

func (p *calcParser) sum() float64 {
	v := p.product()
	for {
		switch p.next("+-") {
		case '+':
			v += p.product()
		case '-':
			v -= p.product()
		default:
			return v
		}
	}
}

func (p *calcParser) product() float64 {
	v := p.unary()
	for {
		op := p.next("*x×/÷%")
		if op == 0 {
			return v
		}
		r := p.unary()
		switch op {
		case '*', 'x', '×':
			v *= r
		case '/', '÷':
			p.bad = p.bad || r == 0
			v /= r
		case '%':
			p.bad = p.bad || r == 0
			v = math.Mod(v, r)
		}
	}
}

// unary binds looser than ^, so -2^2 is -4.
func (p *calcParser) unary() float64 {
	switch p.next("+-") {
	case '+':
		return p.unary()
	case '-':
		return -p.unary()
	}
	return p.power()
}

func (p *calcParser) power() float64 {
	v := p.primary()
	if p.next("^") != 0 {
		// Right to left, so 2^3^2 is 2^9.
		return math.Pow(v, p.unary())
	}
	return v
}

func (p *calcParser) primary() float64 {
	if p.next("(") != 0 {
		v := p.sum()
		p.bad = p.bad || p.next(")") == 0
		return v
	}

	start := p.pos
	for p.pos < len(p.text) && strings.ContainsRune("0123456789.", p.text[p.pos]) {
		p.pos++
	}
	v, err := strconv.ParseFloat(string(p.text[start:p.pos]), 64)
	p.bad = p.bad || err != nil
	return v
}
