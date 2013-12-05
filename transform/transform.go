package transform

import (
	"encoding/xml"
	"io"
	"fmt"
)

// Transform iterates over an XML document passed in via r, calling
// the provided handler for each parsed node.
//
// Any non io.EOF error encountered during the parsing or handling
// stages will be passed to the handler.Error method.  If the
// handler.Error method returns true, then processing will be aborted
// and the error returned.
//
// handler.Flush will be called before Transform returns.
func Transform(r io.Reader, handler Handler) (err error) {
	defer handler.Flush()

	dec := xml.NewDecoder(r)
	for {
		var tok xml.Token
		if tok, err = dec.Token(); err == nil {
			switch node := tok.(type) {
			case xml.StartElement:
				err = handler.StartElement(node)
			case xml.EndElement:
				err = handler.EndElement(node)
			case xml.CharData:
				err = handler.CharData(node)
			case xml.Comment:
				err = handler.Comment(node)
			case xml.Directive:
				err = handler.Directive(node)
			case xml.ProcInst:
				err = handler.ProcInst(node)
			default:
				err = fmt.Errorf("unhandled type: %v %v", node, tok)
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if handler.Error(err) {
				return err
			}
		}
	}

	return
}

// Handler defines methods to handle the possible states provided by
// the XML parser.
type Handler interface {
	//  StartElementwill be called when the parser reports an open tag
	StartElement(xml.StartElement) error

	// EndElement will be called when the parser reports a close tag
	EndElement(xml.EndElement) error

	// CharData will be called when the parser reports XML character data
	CharData(xml.CharData) error

	// Comment will be called when the parser reports an XML comment
	Comment(xml.Comment) error

	// Directive will be called when the parser reports an XML directive
	Directive(xml.Directive) error

	// ProcInst will be called when the parser reports an XML processing instruction
	ProcInst(xml.ProcInst) error

	// Flush indicates a request to flush the underlying io.Writer
	Flush() error

	// Error will be passed any non-io.EOF errors for evaluation
	// or reporting.  If true is returned, the Transform will be
	// aborted and the error returned.
	Error(error) (abort bool)
}
