package eval

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

// maxCalculateLength bounds the text calculate parses. Arithmetic people
// write in a message is far shorter, and it keeps the recursion shallow.
const maxCalculateLength = 200

// calculate returns the value of arithmetic like "(3 * 4) - 2", or nil if the
// text is anything else, like a chat message. The text comes from users, like
// the messages of a counting channel, so it's parsed here instead of by expr,
// which would also run functions and ranges. It knows + - * / % ^, x and × for
// times, ÷, and parentheses. Whole results are ints, so they print without a
// decimal point.
func calculate(text string) any {
	if len(text) > maxCalculateLength {
		return nil
	}

	p := calcParser{text: []rune(text)}
	v, ok := p.sum()
	p.skipSpace()
	if !ok || p.pos != len(p.text) || math.IsInf(v, 0) || math.IsNaN(v) {
		return nil
	}

	if v == math.Trunc(v) && math.Abs(v) < 1<<53 {
		return int(v)
	}
	return v
}

type calcParser struct {
	text []rune
	pos  int
}

func (p *calcParser) skipSpace() {
	for p.pos < len(p.text) && unicode.IsSpace(p.text[p.pos]) {
		p.pos++
	}
}

// next consumes the next rune if it's one of ops.
func (p *calcParser) next(ops string) (rune, bool) {
	p.skipSpace()
	if p.pos < len(p.text) && strings.ContainsRune(ops, p.text[p.pos]) {
		p.pos++
		return p.text[p.pos-1], true
	}
	return 0, false
}

func (p *calcParser) sum() (float64, bool) {
	v, ok := p.product()
	for ok {
		op, found := p.next("+-")
		if !found {
			break
		}
		var r float64
		r, ok = p.product()
		if op == '+' {
			v += r
		} else {
			v -= r
		}
	}
	return v, ok
}

func (p *calcParser) product() (float64, bool) {
	v, ok := p.unary()
	for ok {
		op, found := p.next("*x×/÷%")
		if !found {
			break
		}
		var r float64
		r, ok = p.unary()
		switch {
		case !ok:
		case op == '/' || op == '÷' || op == '%':
			if r == 0 {
				return 0, false
			}
			if op == '%' {
				v = math.Mod(v, r)
			} else {
				v /= r
			}
		default:
			v *= r
		}
	}
	return v, ok
}

// unary binds looser than ^, so -2^2 is -4.
func (p *calcParser) unary() (float64, bool) {
	if op, ok := p.next("+-"); ok {
		v, ok := p.unary()
		if op == '-' {
			v = -v
		}
		return v, ok
	}
	return p.power()
}

func (p *calcParser) power() (float64, bool) {
	v, ok := p.primary()
	if !ok {
		return 0, false
	}
	if _, found := p.next("^"); found {
		// Right to left, so 2^3^2 is 2^9.
		exp, ok := p.unary()
		return math.Pow(v, exp), ok
	}
	return v, true
}

func (p *calcParser) primary() (float64, bool) {
	if _, ok := p.next("("); ok {
		v, ok := p.sum()
		if _, closed := p.next(")"); !ok || !closed {
			return 0, false
		}
		return v, true
	}

	start := p.pos
	for p.pos < len(p.text) && (p.text[p.pos] >= '0' && p.text[p.pos] <= '9' || p.text[p.pos] == '.') {
		p.pos++
	}
	v, err := strconv.ParseFloat(string(p.text[start:p.pos]), 64)
	return v, err == nil
}
