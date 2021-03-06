package transform

import (
	"encoding/xml"
	"github.com/jimrobinson/xml/xmlns"
	"io"
)

// IdentityTransform implements a Handler that writes serialzed XML
// that is semantically, but not necessarily syntactically, equivalent
// to its input.
type IdentityTransform struct {
	w  io.Writer
	ns *xmlns.XmlNamespace
}

func NewIdentityTransform(w io.Writer) *IdentityTransform {
	return &IdentityTransform{
		w:  w,
		ns: xmlns.NewXmlNamespace(),
	}
}

const xmlnsPrefix = "xmlns"
const xmlSpace = "http://www.w3.org/XML/1998/namespace"

var xmlDecl = []byte("xml:")
var xmlnsDecl = []byte("xmlns:")

var startStartElement = []byte("<")
var endStartElement = []byte(">")

var startAttr = []byte("='")
var endAttr = []byte("'")

var colon = []byte(":")
var space = []byte(" ")

func (t *IdentityTransform) StartElement(node xml.StartElement) (err error) {
	t.ns.Push(node)

	t.w.Write(startStartElement)
	if node.Name.Space != "" {
		if p := t.ns.Prefix(node.Name.Space); p != "" {
			t.w.Write([]byte(p))
			t.w.Write(colon)
		}
	}
	t.w.Write([]byte(node.Name.Local))
	for i := range node.Attr {
		attr := node.Attr[i]
		t.w.Write(space)
		if attr.Name.Space != "" {
			if attr.Name.Space == xmlSpace {
				t.w.Write(xmlDecl)
			} else if attr.Name.Space == xmlnsPrefix {
				t.w.Write(xmlnsDecl)
			} else if p := t.ns.Prefix(attr.Name.Space); p != "" {
				t.w.Write([]byte(p))
				t.w.Write(colon)
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
			t.w.Write([]byte(p))
			t.w.Write(colon)
		}
	}
	t.w.Write([]byte(node.Name.Local))
	t.w.Write(endEndElement)

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
	t.w.Write([]byte(node.Target))
	t.w.Write(space)
	t.w.Write(node.Inst)
	t.w.Write(endProcInst)
	return
}

func (t *IdentityTransform) Error(err error) (abort bool) {
	return true
}

func (t *IdentityTransform) Flush() (err error) {
	return nil
}
