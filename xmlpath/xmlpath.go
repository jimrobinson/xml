package xmlpath

import (
	"encoding/xml"
	"fmt"
	"path"

	"github.com/jimrobinson/xml/xmlns"
)

type XmlPath struct {
	ns   *xmlns.XmlNamespace
	path []string
}

func NewXmlPath() *XmlPath {
	return &XmlPath{
		ns:   xmlns.NewXmlNamespace(),
		path: make([]string, 0),
	}
}

func (xp *XmlPath) Push(node xml.StartElement) {
	xp.ns.Push(node)

	var name string
	if prefix := xp.ns.Prefix(node.Name.Space); prefix != "" {
		name = fmt.Sprintf("%s:%s", prefix, node.Name.Local)
	} else {
		name = node.Name.Local
	}

	xp.path = append(xp.path, name)
}

func (xp *XmlPath) Pop() {
	if len(xp.path) == 0 {
		return
	}
	xp.ns.Pop()
	xp.path = xp.path[0 : len(xp.path)-1]
}

func (xp *XmlPath) Peek() string {
	if len(xp.path) == 0 {
		return "/"
	}
	return xp.path[len(xp.path)-1]
}

func (xp *XmlPath) String() string {
	return fmt.Sprintf("/%s", path.Join(xp.path...))
}

func (xp *XmlPath) XmlnsCheck(node xml.StartElement) error {
	return xp.ns.Check(node)
}

func (xp *XmlPath) XmlNsInScope() (xmlns []xml.Attr) {
	return xp.ns.InScopeXmlns()
}

func (xp *XmlPath) XmlNsPrefix(uri string) string {
	return xp.ns.Prefix(uri)
}
