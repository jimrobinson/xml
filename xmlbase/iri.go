// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlbase

import (
	"bytes"
	"net/url"
	"strconv"
)

type encoding int

const (
	encodePath encoding = 1 + iota
	encodeUserPassword
	encodeQueryComponent
	encodeFragment
)

type IRI struct {
	*url.URL
}

func NewIRI(rawurl string) (iri *IRI, err error) {
	var u *url.URL
	u, err = url.Parse(rawurl)
	if err == nil {
		iri = &IRI{URL: u}
	}
	return
}

func (u *IRI) String() (result string, err error) {
	var buf bytes.Buffer
	if u.Scheme != "" {
		buf.WriteString(u.Scheme)
		buf.WriteByte(':')
	}
	if u.Opaque != "" {
		buf.WriteString(u.Opaque)
	} else {
		if u.Scheme != "" || u.Host != "" || u.User != nil {
			buf.WriteString("//")
			if u := u.User; u != nil {
				buf.WriteString(u.String())
				buf.WriteByte('@')
			}
			if h := u.Host; h != "" {
				buf.WriteString(h)
			}
		}
		s, err := unescape(u.Path, encodePath)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
	}
	if u.RawQuery != "" {
		buf.WriteByte('?')
		buf.WriteString(u.RawQuery)
	}
	if u.Fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(escape(u.Fragment, encodeFragment))
	}
	return buf.String(), nil
}

func (iri *IRI) ResolveReference(ref *IRI) *IRI {
	// if ref has only #Fragment then it is a no-op
	if ref.Scheme == "" && ref.Opaque == "" && ref.User == nil && ref.Host == "" && ref.Path == "" && ref.RawQuery == "" {
		return iri
	}
	return &IRI{URL: iri.URL.ResolveReference(ref.URL)}
}

type EscapeError string

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

// Return true if the specified character should be escaped when
// appearing in a URL string, according to RFC 3986.
// When 'all' is true the full range of reserved characters are matched.
func shouldEscape(c byte, mode encoding) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)
		// Different sections of the URL allow a few of
		// the reserved characters to appear unescaped.
		switch mode {
		case encodePath: // §3.3
			// The RFC allows : @ & = + $ but saves / ; , for assigning
			// meaning to individual path segments. This package
			// only manipulates the path as a whole, so we allow those
			// last two as well. That leaves only ? to escape.
			return c == '?'

		case encodeUserPassword: // §3.2.2
			// The RFC allows ; : & = + $ , in userinfo, so we must escape only @ and /.
			// The parsing of userinfo treats : as special so we must escape that too.
			return c == '@' || c == '/' || c == ':'

		case encodeQueryComponent: // §3.4
			// The RFC reserves (so we must escape) everything.
			return true

		case encodeFragment: // §4.1
			// The RFC text is silent but the grammar allows
			// everything, so escape nothing.
			return false
		}
	}

	// Everything else must be escaped.
	return true
}

func unescape(s string, mode encoding) (string, error) {
	// Count %, check that they're well-formed.
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[0:3]
				}
				return "", EscapeError(s)
			}
			i += 3
		case '+':
			hasPlus = mode == encodeQueryComponent
			i++
		default:
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s, nil
	}

	t := make([]byte, len(s)-2*n)
	j := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			t[j] = unhex(s[i+1])<<4 | unhex(s[i+2])
			j++
			i += 3
		case '+':
			if mode == encodeQueryComponent {
				t[j] = ' '
			} else {
				t[j] = '+'
			}
			j++
			i++
		default:
			t[j] = s[i]
			j++
			i++
		}
	}
	return string(t), nil
}

func escape(s string, mode encoding) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, mode) {
			if c == ' ' && mode == encodeQueryComponent {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && mode == encodeQueryComponent:
			t[j] = '+'
			j++
		case shouldEscape(c, mode):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
