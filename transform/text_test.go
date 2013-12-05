package transform

import (
	"bytes"
	"testing"
)

type Char struct {
	Input []byte
	Output []byte
}

type TypedCharTest struct {
	nodeType NodeType
	Chars []Char
}

var escapeTextTests = []TypedCharTest{
	{
		AttrValue,
		[]Char{
			{[]byte(`"`), esc_quot},
			{[]byte(`'`), esc_apos},
			{[]byte(`&`), esc_amp},
			{[]byte(`<`), esc_lt},
			{[]byte(`>`), []byte(`>`)},
			{[]byte("\t"), esc_tab},
			{[]byte("\n"), esc_nl},
			{[]byte("\r"), esc_cr},
			{[]byte(`a`), []byte(`a`)},
			{[]byte(`é`), []byte(`é`)},
			{[]byte("\u0011"), []byte("\uFFFD")},
		},
	},
	{
		CharData,
		[]Char{
			{[]byte(`"`), []byte(`"`)},
			{[]byte(`'`) , []byte(`'`)},
			{[]byte(`&`), esc_amp},
			{[]byte(`<`), esc_lt},
			{[]byte(`>`), esc_gt},
			{[]byte("\t"), []byte("\t")},
			{[]byte("\n"), []byte("\n")},
			{[]byte("\r"), esc_cr},
			{[]byte(`a`), []byte(`a`)},
			{[]byte(`é`), []byte(`é`)},
			{[]byte("\u0011"), []byte("\uFFFD")},
		},
	},
}

func TestEscapeNodeValue(t *testing.T) {
	for i, v := range escapeTextTests {
		for j, c := range v.Chars {
			w := &bytes.Buffer{}
			s := c.Input
			n := v.nodeType
			EscapeNodeValue(w, s, n)
			if 0 != bytes.Compare(w.Bytes(), c.Output) {
				t.Error(i, j, c.Input, c.Output, w.Bytes())
			}
		}
	}
}
