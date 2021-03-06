// XMLBase will track the active xml:base value for a given point in
// the XML tree.
//
// For every xml.StartElement node encountered, pass the node to the
// Push function before completing any other processing that requires
// the xml:base.
//
// For every xml.EndElement node encountered, call the Pop function
// after completing any other processing requiring the current
// xml:base.
package xmlbase

import (
	"golang.org/x/net/html"
	"encoding/xml"
	"strings"
)

type XmlBase struct {
	baseUri []*IRI
	depth   []int
}

func NewXmlBase(baseUri string) (xb *XmlBase, err error) {
	var u *IRI
	u, err = NewIRI(baseUri)
	if err != nil {
		return
	}
	xb = new(XmlBase)
	xb.baseUri = append(xb.baseUri, u)
	xb.depth = append(xb.depth, 1)
	return
}

const xmlBaseSpace = "http://www.w3.org/XML/1998/namespace"
const xmlBaseLocal = "base"

// Push adds xml:base from xml.StartElement to the stack
func (xb *XmlBase) Push(node xml.StartElement) (err error) {
	var rawurl string
	var exists bool
	for _, attr := range node.Attr {
		if attr.Name.Space == xmlBaseSpace && attr.Name.Local == xmlBaseLocal {
			rawurl = attr.Value
			exists = true
			break
		}
	}
	return xb.push(rawurl, exists)
}

// PushHTML adds xml:base from html.Token to the stack
func (xb *XmlBase) PushHTML(tok html.Token) (err error) {
	var rawurl string
	var exists bool
	for _, attr := range tok.Attr {
		if attr.Key == "xml:base" {
			rawurl = attr.Val
			exists = true
			break
		}
		if attr.Namespace == xmlBaseSpace && (strings.HasSuffix(attr.Key, ":base") || attr.Key == "base") {
			rawurl = attr.Val
			exists = true
			break
		}
	}
	return xb.push(rawurl, exists)
}

func (xb *XmlBase) push(rawurl string, exists bool) (err error) {
	n := len(xb.baseUri) - 1
	if !exists {
		xb.depth[n]++
		return
	}

	var u *IRI
	u, err = NewIRI(rawurl)
	if err != nil {
		return
	}

	if !u.IsAbs() {
		u = xb.baseUri[n].ResolveReference(u)
	}

	x1, err1 := xb.baseUri[n].String()
	if err1 != nil {
		return err1
	}
	x2, err2 := u.String()
	if err != nil {
		return err2
	}
	if x1 == x2 {
		xb.depth[n]++
		return
	}

	xb.baseUri = append(xb.baseUri, u)
	xb.depth = append(xb.depth, 1)
	return
}

// Pop removes the latest xml:base from the stack
func (xb *XmlBase) Pop() {
	n := len(xb.baseUri) - 1
	if n <= 0 {
		if n == 0 && xb.depth[n] > 0 {
			xb.depth[n]--
		}
		return
	}
	xb.depth[n]--
	if xb.depth[n] <= 0 {
		xb.baseUri = xb.baseUri[0:n]
		xb.depth = xb.depth[0:n]
	}
	return
}

// Resolve returns the resolved version of rawurl based on the current xml:base URL
func (xb *XmlBase) Resolve(rawurl string) (iri string, err error) {
	var u *IRI
	u, err = NewIRI(rawurl)
	if err != nil {
		return
	}
	if !u.IsAbs() {
		n := len(xb.baseUri) - 1
		u = xb.baseUri[n].ResolveReference(u)
	}
	return u.String()
}

// URL returns the current xml:base URL
func (xb *XmlBase) URL() *IRI {
	return xb.baseUri[len(xb.baseUri)-1]
}
