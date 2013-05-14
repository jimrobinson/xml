package transform

import (
	"bufio"
	"encoding/xml"
	"github.com/jimrobinson/xml/xmlbase"
	"github.com/jimrobinson/xml/xmlns"
	"io"
)

// IdentityTransform implements a Handler that writes serialzed XML
// that is semantically, but not necessarily syntactically, equivalent
// to its input.
type IdentityTransform struct {
	w    *bufio.Writer
	ns   *xmlns.XmlNamespace
	base *xmlbase.XmlBase
}

func NewIdentityTransform(w io.Writer, baseUri string) (t *IdentityTransform, err error) {
	var bufw *bufio.Writer
	var ok bool

	if bufw, ok = w.(*bufio.Writer); !ok {
		bufw = bufio.NewWriter(w)
	}

	var ns *xmlns.XmlNamespace
	ns = xmlns.NewXmlNamespace()

	var base *xmlbase.XmlBase
	base, err = xmlbase.NewXmlBase(baseUri)
	if err != nil {
		return
	}

	t = &IdentityTransform{
		w:    bufw,
		ns:   ns,
		base: base,
	}

	return
}

var startStartElement = []byte("<")
var endStartElement = []byte(">")

var startAttr = []byte(`='`)
var endAttr = []byte(`'`)

func (t *IdentityTransform) StartElement(node xml.StartElement) (err error) {
	t.ns.Push(node)
	err = t.base.Push(node)
	if err != nil {
		return
	}

	t.w.Write(startStartElement)
	if node.Name.Space != "" {
		if p := t.ns.Prefix(node.Name.Space); p != "" {
			t.w.WriteString(p)
			t.w.WriteByte(':')
		} else {
			t.w.WriteString(node.Name.Space)
			t.w.WriteByte(':')
		}
	}
	t.w.WriteString(node.Name.Local)
	for i := range node.Attr {
		attr := node.Attr[i]
		t.w.WriteByte(' ')
		if attr.Name.Space != "" {
			if p := t.ns.Prefix(attr.Name.Space); p != "" {
				t.w.WriteString(p)
				t.w.WriteByte(':')
			} else {
				t.w.WriteString(attr.Name.Space)
				t.w.WriteByte(':')
			}
		}
		t.w.Write([]byte(attr.Name.Local))
		t.w.Write(startAttr)
		if err = EscapeNodeValue(t.w, []byte(attr.Value), AttrValue); err != nil {
			return
		}
		t.w.Write(endAttr)
	}
	t.w.Write(endStartElement)
	return
}

var startEndElement = []byte("</")
var endEndElement = []byte(">")

func (t *IdentityTransform) EndElement(node xml.EndElement) (err error) {
	t.w.Write(startEndElement)
	if node.Name.Space != "" {
		if p := t.ns.Prefix(node.Name.Space); p != "" {
			t.w.WriteString(p)
			t.w.WriteByte(':')
		} else {
			t.w.WriteString(node.Name.Space)
			t.w.WriteByte(':')
		}
	}
	t.w.WriteString(node.Name.Local)
	t.w.Write(endEndElement)

	t.base.Pop()
	t.ns.Pop()
	return
}

func (t *IdentityTransform) CharData(node xml.CharData) (err error) {
	return EscapeNodeValue(t.w, node, CharData)
}

var startComment = []byte("<!--")
var endComment = []byte("-->")

func (t *IdentityTransform) Comment(node xml.Comment) (err error) {
	t.w.Write(startComment)
	t.w.Write(node)
	t.w.Write(endComment)
	return
}

var startDirective = []byte("<!")
var endDirective = []byte(">")

func (t *IdentityTransform) Directive(node xml.Directive) (err error) {
	t.w.Write(startDirective)
	t.w.Write(node)
	t.w.Write(endDirective)
	return
}

var startProcInst = []byte("<?")
var endProcInst = []byte("?>")

func (t *IdentityTransform) ProcInst(node xml.ProcInst) (err error) {
	t.w.Write(startProcInst)
	t.w.WriteString(node.Target)
	t.w.WriteByte(' ')
	t.w.Write(node.Inst)
	t.w.Write(endProcInst)
	return
}

func (t *IdentityTransform) Error(err error) (abort bool) {
	return true
}

func (t *IdentityTransform) Flush() (err error) {
	return t.w.Flush()
}