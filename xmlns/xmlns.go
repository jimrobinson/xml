package xmlns

import (
	"encoding/xml"
)

const xml_namespace = "http://www.w3.org/XML/1998/namespace"
const xml_prefix = "xml"

// XmlNamespace tracks the mapping of XML namespaces in an XML tree.
// For every xml.StartElement node encountered, pass the node to the
// Push function.  For every xml.EndElement node encountered, call the
// Pop function.
type XmlNamespace struct {
	Stack []*Mapping
}

// Prefix maps a single namespace prefix to a uri
type Prefix map[string]string

// Uri maps a single namespace uri to one or more prefixes
type Uri map[string][]string

// Mapping tracks the namespaces defined for a particular depth of the XML parse tree
type Mapping struct {
	Prefix Prefix // mapping of prefix to namespaces
	Uri    Uri    // mapping of namespaces to one or more prefixes
	depth  int    // mapping is good for this many EndElement nodes
}

func NewXmlNamespace() *XmlNamespace {
	return &XmlNamespace{}
}

// Push adds namespace mappings onto the stack
func (ns *XmlNamespace) Push(node xml.StartElement) {
	prefix := make(Prefix)
	uri := make(Uri)
	for _, attr := range node.Attr {
		if attr.Name.Space != "xmlns" {
			continue
		}

		if _, ok := prefix[attr.Name.Local]; !ok {
			uri[attr.Value] = append(uri[attr.Value], attr.Name.Local)
		}
		prefix[attr.Name.Local] = attr.Value
	}
	if len(prefix) == 0 {
		n := len(ns.Stack) - 1
		if n < 0 {
			ns.Stack = append(ns.Stack, &Mapping{Prefix: prefix, Uri: uri, depth: 1})
			return
		}
		ns.Stack[n].depth++
		return
	}
	ns.Stack = append(ns.Stack, &Mapping{Prefix: prefix, Uri: uri, depth: 1})
}

// Pop removes namespace mappings from the stack
func (ns *XmlNamespace) Pop() {
	n := len(ns.Stack) - 1
	if n < 0 {
		return
	}
	ns.Stack[n].depth--
	if ns.Stack[n].depth <= 0 {
		ns.Stack = ns.Stack[0:n]
	}
}

// InScope returns a Mapping of namespaces that are currently in scope, or nil
func (ns *XmlNamespace) InScope() *Mapping {
	n := len(ns.Stack) - 1
	if n < 0 {
		return nil
	}
	scope := &Mapping{
		Prefix: make(Prefix),
		Uri:    make(Uri),
	}
	for i := n; i >= 0; i-- {
		for k, v := range ns.Stack[i].Prefix {
			if _, ok := scope.Prefix[k]; ok {
				continue
			}
			scope.Prefix[k] = v
			scope.Uri[v] = append(scope.Uri[v], k)
		}
	}
	return scope
}

// InScopeXmlns returns a slice of xmlns attributes for namespaces currently in scope.
func (ns *XmlNamespace) InScopeXmlns() (xmlns []xml.Attr) {
	m := ns.InScope()
	if m == nil {
		return
	}
	xmlns = make([]xml.Attr, len(m.Prefix))
	for k, v := range m.Prefix {
		xmlns = append(xmlns, xml.Attr{Name: xml.Name{Space: "xmlns", Local: k}, Value: v})
	}
	return
}

// Prefix returns the current prefix for a namespace uri, or the empty
// string.  If more than one prefix was mapped to the uri, the first
// prefix mapped in the closest element to the current location will
// be returned.
func (ns *XmlNamespace) Prefix(uri string) string {
	if uri == xml_namespace {
		return xml_prefix
	}

	n := len(ns.Stack) - 1
	if n < 0 {
		return ""
	}
	for i := n; i >= 0; i-- {
		if prefix, ok := ns.Stack[i].Uri[uri]; ok {
			return prefix[0]
		}
	}
	return ""
}
