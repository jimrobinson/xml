xml
===

A collection of helper libraries for use with
[encoding/xml](http://golang.org/pkg/encoding/xml/):

- transform: Facilitate a streaming transformation of XML
- xmlbase: Track current xml:base as an XML document is parsed.

If the XML you are processing is already mapped to a go structure, it
makes more sense to just use the existing
[Marshal](http://golang.org/pkg/encoding/xml/#Marshal) and
[Unmarshal](http://golang.org/pkg/encoding/xml/#Unmarshal) facilities
in the standard xml processing library.

If aren't mapping the XML, if you are dealing with a stream of data
and you want to introduce small changes to that stream, then transform
may be of some help.

The transform library sits on top of the standard go xml parser,
allowing the user to introduce changes using an event handler.  A
default event handler that provides a basic identity transform is
provided.

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
		<img alt="demo" src="demo.gif" />
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

produces the output, almost identical except that xhtml element href
and src attributes have been fully qualified using the provided base
URI (http://example.com/):

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
		<img alt="demo" src="http://example.com/demo.gif" />
	    </p>
	  </body>
	</html>
