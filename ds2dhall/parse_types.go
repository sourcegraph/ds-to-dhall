package ds2dhall

import (
	"bytes"
	"fmt"
	"strings"
	"text/scanner"
	"unicode"
)

type parser struct {
	scnr   *scanner.Scanner
	peeked string
	res    map[string]string
	err    error
	key    string
	value  string
}

func (p *parser) consume() string {
	if p.peeked != "" {
		r := p.peeked
		p.peeked = ""
		return r
	}
	tk := p.scnr.Scan()
	if tk == scanner.EOF {
		return ""
	}
	return p.scnr.TokenText()
}

func (p *parser) peek() string {
	if p.peeked == "" {
		tk := p.scnr.Scan()
		if tk == scanner.EOF {
			return ""
		}
		p.peeked = p.scnr.TokenText()
	}
	return p.peeked
}

func (p *parser) parseToken(token string) {
	if p.err != nil {
		return
	}

	t := p.consume()
	if t == "" {
		p.err = fmt.Errorf("unexpected EOF")
	}
	if t != token {
		p.err = fmt.Errorf("expected %s, got %s", token, t)
	}
}

func (p *parser) parseKey() {
	if p.err != nil {
		return
	}

	t := p.consume()
	if t == "" {
		p.err = fmt.Errorf("unexpected EOF")
	}
	p.key = t
}

func (p *parser) parseValue() {
	if p.err != nil {
		return
	}

	t := p.consume()
	if t == "" {
		p.err = fmt.Errorf("unexpected EOF")
	}
	p.value = t
}

func (p *parser) parseOptionalSha256() {
	if p.err != nil {
		return
	}

	t := p.peek()
	if t == "" || !strings.HasPrefix(t, "sha256:") {
		return
	}
	p.consume()
}

func (p *parser) parseRecord() {
	if p.err != nil {
		return
	}

	p.parseToken("{")
	p.parseEntryList()
	p.parseToken("}")
}

func (p *parser) parseEntryList() {
	if p.err != nil {
		return
	}

	p.parseOptionalEntry()
	for p.parseOptionalCommaEntry() {
	}
}

func (p *parser) parseOptionalEntry() {
	if p.err != nil {
		return
	}

	t := p.peek()
	if t == "" || t == "}" {
		return
	}
	p.parseEntry()
}

func (p *parser) parseOptionalCommaEntry() bool {
	if p.err != nil {
		return false
	}

	t := p.peek()
	if t != "," {
		return false
	}

	p.parseToken(",")
	p.parseEntry()
	return true
}

func (p *parser) parseEntry() {
	if p.err != nil {
		return
	}

	p.parseKey()
	p.parseToken("=")
	p.parseValue()
	p.parseOptionalSha256()

	if p.err == nil {
		p.res[p.key] = p.value
		p.key = ""
		p.value = ""
	}
}

func parseTypes(data []byte) (map[string]string, error) {
	var s scanner.Scanner
	s.Init(bytes.NewReader(data))
	s.Filename = "types.dhall"
	s.IsIdentRune = func(ch rune, i int) bool {
		return ch == '.' || (ch == ':' || ch == '/') && i > 0 || unicode.IsLetter(ch) || (unicode.IsDigit(ch) || unicode.IsPunct(ch)) && i > 0
	}

	p := &parser{
		scnr: &s,
		res:  make(map[string]string),
	}

	p.parseRecord()

	return p.res, p.err
}
