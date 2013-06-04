package xmlns

import (
	"encoding/xml"
	"fmt"
)

const xmlPrefix = "xml"
const xmlnsPrefix = "xmlns"
const xmlnsSpace = "http://www.w3.org/XML/1998/namespace"

// XmlNamespace tracks the mapping of XML namespaces in an XML tree.
// For every xml.StartElement node encountered, pass the node to the
// Push function.  For every xml.EndElement node encountered, call the
// Pop function.
type XmlNamespace struct {
	Stack []*Mapping
	Scope []*Mapping // in-scope namespaces
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

func copyScope(n, o *Mapping) {
	for k, v := range o.Prefix {
		n.Prefix[k] = v
	}
	for k, v := range o.Uri {
		n.Uri[k] = v
	}
	n.depth = o.depth
}

// Push adds namespace mappings onto the stack
func (ns *XmlNamespace) Push(node xml.StartElement) {

	// the current element namespace declarations
	stack := &Mapping{
		Prefix: make(Prefix),
		Uri:    make(Uri),
	}

	for _, attr := range node.Attr {
		if attr.Name.Space != xmlnsPrefix {
			continue
		}
		if _, ok := stack.Prefix[attr.Name.Local]; !ok {
			stack.Uri[attr.Value] = append(stack.Uri[attr.Value], attr.Name.Local)
		}
		stack.Prefix[attr.Name.Local] = attr.Value
	}
	if len(stack.Prefix) == 0 {
		// no declarations were in this node, increment depth
		n := len(ns.Stack) - 1
		if n < 0 {
			stack.depth = 1
			ns.Stack = append(ns.Stack, stack)
			ns.Scope = append(ns.Scope, stack)
			return
		}
		ns.Stack[n].depth++
		ns.Scope[n].depth++
		return
	}

	// declarations were found, push onto the stack
	stack.depth = 1
	ns.Stack = append(ns.Stack, stack)

	// new scope by merging old scope with current stack
	scope := &Mapping{
		Prefix: make(Prefix),
		Uri:    make(Uri),
	}
	if len(ns.Scope) > 0 {
		copyScope(scope, ns.Scope[len(ns.Scope)-1])
	}
	copyScope(scope, stack)
	ns.Scope = append(ns.Scope, scope)
}

// Check examines a node for missing namespace mappings on the node.
// If an unmapped namespace is discovered an error will be returned.
// Push needs to be called before Check.
func (ns *XmlNamespace) Check(node xml.StartElement) error {
	m := ns.InScope()
	if node.Name.Space != "" && node.Name.Space != xmlPrefix && node.Name.Space != xmlnsSpace && node.Name.Space != xmlnsPrefix {
		if _, ok := m.Uri[node.Name.Space]; !ok {
			return fmt.Errorf("unmapped namespace prefix: %s", node.Name.Space)
		}
	}
	for _, attr := range node.Attr {
		if attr.Name.Space != "" && attr.Name.Space != xmlPrefix && attr.Name.Space != xmlnsSpace && attr.Name.Space != xmlnsPrefix {
			if _, ok := m.Uri[attr.Name.Space]; !ok {
				return fmt.Errorf("unmapped namespace prefix: %s", attr.Name.Space)
			}
		}
	}
	return nil
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
	ns.Scope[n].depth--
	if ns.Scope[n].depth <= 0 {
		ns.Scope = ns.Scope[0:n]
	}
}

// InScope returns a Mapping of namespaces that are currently in scope, or nil
func (ns *XmlNamespace) InScope() *Mapping {
	n := len(ns.Stack) - 1
	if n < 0 {
		return nil
	}
	return ns.Scope[n]
}

// InScopeXmlns returns a slice of xmlns attributes for namespaces currently in scope.
func (ns *XmlNamespace) InScopeXmlns() (xmlns []xml.Attr) {
	m := ns.InScope()
	if m == nil {
		return
	}
	xmlns = make([]xml.Attr, len(m.Prefix))
	for k, v := range m.Prefix {
		xmlns = append(xmlns, xml.Attr{Name: xml.Name{Space: xmlnsPrefix, Local: k}, Value: v})
	}
	return
}

// Prefix returns the current prefix for a namespace uri, or the empty
// string.  If more than one prefix was mapped to the uri, the first
// prefix mapped in the closest element to the current location will
// be returned.
func (ns *XmlNamespace) Prefix(uri string) string {
	if uri == xmlPrefix {
		return xmlPrefix
	}

	n := len(ns.Scope) - 1
	if n < 0 {
		return ""
	}
	for i := n; i >= 0; i-- {
		if prefix, ok := ns.Scope[i].Uri[uri]; ok {
			return prefix[0]
		}
	}
	return ""
}
