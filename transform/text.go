package transform

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// References:
//    Annotated XML spec: http://www.xml.com/axml/testaxml.htm
//    XML name spaces: http://www.w3.org/TR/REC-xml-names/

// TODO(rsc):
//	Test error handling.

import (
	"io"
	"unicode/utf8"
)

type NodeType int

const (
	AttrValue NodeType = iota
	CharData
)

var (
	esc_quot = []byte("&#34;") // shorter than "&quot;"
	esc_apos = []byte("&#39;") // shorter than "&apos;"
	esc_amp  = []byte("&amp;")
	esc_lt   = []byte("&lt;")
	esc_gt   = []byte("&gt;")
	esc_tab  = []byte("&#x9;")
	esc_nl   = []byte("&#xA;")
	esc_cr   = []byte("&#xD;")
	esc_fffd = []byte("\uFFFD") // Unicode replacement character
)

// Decide whether the given rune is in the XML Character Range, per
// the Char production of http://www.xml.com/axml/testaxml.htm,
// Section 2.2 Characters.
func isInCharacterRange(r rune) (inrange bool) {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xDF77 ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}

// EscapeNodeValue writes to w the properly escaped XML equivalent of
// the plain text data s for node type t.
func EscapeNodeValue(w io.Writer, s []byte, t NodeType) error {
	var esc []byte
	last := 0

	switch t {
	case AttrValue:
		for i := 0; i < len(s); {
			r, width := utf8.DecodeRune(s[i:])
			i += width
			switch r {
			case '"':
				esc = esc_quot
			case '\'':
				esc = esc_apos
			case '&':
				esc = esc_amp
			case '<':
				esc = esc_lt
			case '\t':
				esc = esc_tab
			case '\n':
				esc = esc_nl
			case '\r':
				esc = esc_cr
			default:
				if !isInCharacterRange(r) {
					esc = esc_fffd
					break
				}
				continue
			}
			if _, err := w.Write(s[last : i-width]); err != nil {
				return err
			}
			if _, err := w.Write(esc); err != nil {
				return err
			}
			last = i
		}
	case CharData:
		for i := 0; i < len(s); {
			r, width := utf8.DecodeRune(s[i:])
			i += width
			switch r {
			case '&':
				esc = esc_amp
			case '<':
				esc = esc_lt
			case '>':
				esc = esc_gt
			case '\r':
				esc = esc_cr
			default:
				if !isInCharacterRange(r) {
					esc = esc_fffd
					break
				}
				continue
			}
			if _, err := w.Write(s[last : i-width]); err != nil {
				return err
			}
			if _, err := w.Write(esc); err != nil {
				return err
			}
			last = i
		}
	}
	if _, err := w.Write(s[last:]); err != nil {
		return err
	}
	return nil
}
