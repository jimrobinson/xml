xml
===

A collection of helper libraries for use with
[encoding/xml](http://golang.org/pkg/encoding/xml/):

- xmlbase: Track current xml:base as an XML document is parsed.
- transform: Facilitate a streaming transformation of XML

Installation
------------

	$ go get github.com/jimrobinson/xml/transform
	$ go get github.com/jimrobinson/xml/xmlbase

Example
-------

	package main

	import (
		"encoding/xml"
		"github.com/jimrobinson/xml/transform"
		"github.com/jimrobinson/xml/xmlbase"
		"io"
		"log"
		"os"
		"strings"
	)

	// ExampleHandler expands xhtml href and src attribute values to fully
	// qualified urls
	type ExampleHandler struct {
		*transform.IdentityTransform
		base *xmlbase.XmlBase
	}

	func NewHandler(w io.Writer, baseUri string) (h *ExampleHandler, err error) {
		var base *xmlbase.XmlBase
		base, err = xmlbase.NewXmlBase(baseUri)
		if err != nil {
			return
		}

		h = &ExampleHandler{
			IdentityTransform: transform.NewIdentityTransform(w),
			base:              base}
		return
	}

	func (h *ExampleHandler) StartElement(node xml.StartElement) (err error) {
		h.base.Push(node)
		if node.Name.Space == "http://www.w3.org/1999/xhtml" {
			for i, attr := range node.Attr {
				if attr.Name.Space == "" && (attr.Name.Local == "href" || attr.Name.Local == "src") {
					node.Attr[i].Value, err = h.base.Resolve(attr.Value)
					if err != nil {
						return
					}
				}
			}
		}
		return h.IdentityTransform.StartElement(node)
	}

	func (h *ExampleHandler) EndElement(node xml.EndElement) error {
		h.base.Pop()
		return h.IdentityTransform.EndElement(node)
	}

	var sampleXml = `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
	                      "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml">
	    <head>
	        <title>Example</title>
	    </head>
	    <body>
	      <p>
	        <a href="value">A link</a>
	      </p>
	    </body>
	</html>`

	func main() {
		h, err := NewHandler(os.Stdout, "http://example.com/")
		if err != nil {
			log.Fatal(err)
		}
		err = transform.Transform(strings.NewReader(sampleXml), h)
		if err != nil {
			log.Fatal(err)
		}
	}

produces the output:

	<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml">
	  <head>
	    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	    <title>Example</title>
	  </head>
	  <body>
	    <p>
	      <a href="http://example.com/value">A link</a>
	    </p>
	  </body>
	</html>
