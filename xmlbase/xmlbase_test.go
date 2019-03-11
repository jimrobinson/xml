package xmlbase

import (
	"golang.org/x/net/html"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
)

const verbose = false

type XmlBaseTestUri struct {
	xml  xml.Name
	html html.Attribute
	iri  string
}

type XmlBaseTest struct {
	example string
	resolve []XmlBaseTestUri
}

var xmlBaseTests = []XmlBaseTest{
	XmlBaseTest{
		example: `
			<html xmlns="http://www.w3.org/1999/xhtml"
				xml:base="http://example.org/today/"
				xmlns:xlink="http://www.w3.org/1999/xlink">
				<head>
					<title>Virtual Library</title>
				</head>
				<body>
					<paragraph>See <link xlink:type="simple" xlink:href="new.xml">what's new</link>!</paragraph>
					<paragraph>Check out the hot picks of the day!</paragraph>
					<olist xml:base="/hotpicks/">
						<item>
							<link xlink:type="simple" xlink:href="pick1.xml">Hot Pick #1</link>
						</item>
						<item>
							<link xlink:type="simple" xlink:href="pick2.xml">Hot Pick #2</link>
						</item>
						<item>
							<link xlink:type="simple" xlink:href="pick3.xml">Hot Pick #3</link>
						</item>
					</olist>
				</body>
			</html>`,
		resolve: []XmlBaseTestUri{
			{xml: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, html: html.Attribute{Key: "xlink:href"}, iri: "http://example.org/today/new.xml"},
			{xml: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, html: html.Attribute{Key: "xlink:href"}, iri: "http://example.org/hotpicks/pick1.xml"},
			{xml: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, html: html.Attribute{Key: "xlink:href"}, iri: "http://example.org/hotpicks/pick2.xml"},
			{xml: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, html: html.Attribute{Key: "xlink:href"}, iri: "http://example.org/hotpicks/pick3.xml"},
		},
	},
	XmlBaseTest{
		example: `<e1 xml:base="http://example.org/wine/"><e2 xml:base="rosé"/></e1>`,
		resolve: []XmlBaseTestUri{
			{xml: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, html: html.Attribute{Key: "xml:base"}, iri: "http://example.org/wine/"},
			{xml: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, html: html.Attribute{Key: "xml:base"}, iri: "http://example.org/wine/rosé"},
		},
	},
	XmlBaseTest{
		example: `<elt xml:base="http://www.example.org/~Dürst/"/>`,
		resolve: []XmlBaseTestUri{
			{xml: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, html: html.Attribute{Key: "xml:base"}, iri: "http://www.example.org/~Dürst/"},
		},
	},
	XmlBaseTest{
		example: `<outer xml:base="http://www.example.org/one/two"> <inner xml:base=""/> </outer>`,
		resolve: []XmlBaseTestUri{
			{xml: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, html: html.Attribute{Key: "xml:base"}, iri: "http://www.example.org/one/two"},
			{xml: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, html: html.Attribute{Key: "xml:base"}, iri: "http://www.example.org/one/two"},
		},
	},
	XmlBaseTest{
		example: `<outer xml:base="http://www.example.org/one/two"> <inner xml:base="#frag"/> </outer>`,
		resolve: []XmlBaseTestUri{
			{xml: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, html: html.Attribute{Key: "xml:base"}, iri: "http://www.example.org/one/two"},
			{xml: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, html: html.Attribute{Key: "xml:base"}, iri: "http://www.example.org/one/two"},
		},
	},
}

func TestXMLBasePush(t *testing.T) {
	for i, v := range xmlBaseTests {
		xmlbase, err := NewXmlBase("")
		if err != nil {
			t.Fatal(i, err)
		}

		if verbose {
			fmt.Println(i, "created", xmlbase.baseUri, xmlbase.depth)
		}

		dec := xml.NewDecoder(strings.NewReader(v.example))
		r := 0
		for {
			tok, err := dec.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(i, err)
			}
			switch node := tok.(type) {
			case xml.StartElement:
				xmlbase.Push(node)
				if verbose {
					fmt.Println(i, "pushed", xmlbase.baseUri, xmlbase.depth)
				}
				for _, attr := range node.Attr {
					if attr.Name.Space == v.resolve[r].xml.Space && attr.Name.Local == v.resolve[r].xml.Local {
						if verbose {
							fmt.Println(i, "verify", attr, v.resolve[r].iri)
						}

						iri, err := xmlbase.Resolve(attr.Value)
						if err != nil {
							t.Fatal(i, r, err)
						}

						if iri != v.resolve[r].iri {
							t.Fatalf("%d %d expected '%s', got '%s'", i, r, v.resolve[r].iri, iri)
						}
						r++
					}
				}
			case xml.EndElement:
				xmlbase.Pop()
				if verbose {
					fmt.Println(i, "popped", xmlbase.baseUri, xmlbase.depth)
				}
			}
		}
	}
}

func TestXMLBasePushHTML(t *testing.T) {
	for i, v := range xmlBaseTests {
		xmlbase, err := NewXmlBase("")
		if err != nil {
			t.Fatal(i, err)
		}

		if verbose {
			fmt.Println(i, "created", xmlbase.baseUri, xmlbase.depth)
		}

		z := html.NewTokenizer(strings.NewReader(v.example))
		r := 0
		for {
			tt := z.Next()
			switch tt {
			case html.ErrorToken:
				err = z.Err()
				if err == io.EOF {
					return
				}
				t.Fatal(i, err)
			case html.StartTagToken:
				node := z.Token()
				xmlbase.PushHTML(node)
				if verbose {
					fmt.Println(i, "pushed", xmlbase.baseUri, xmlbase.depth)
				}
				for _, attr := range node.Attr {
					if attr.Key == v.resolve[r].html.Key {
						if verbose {
							fmt.Println(i, "verify", attr, v.resolve[r].iri)
						}

						iri, err := xmlbase.Resolve(attr.Val)
						if err != nil {
							t.Fatal(i, r, err)
						}

						if iri != v.resolve[r].iri {
							t.Fatalf("%d %d expected '%s', got '%s'", i, r, v.resolve[r].iri, iri)
						}
						r++
					}
				}

			case html.EndTagToken:
				xmlbase.Pop()
				if verbose {
					fmt.Println(i, "popped", xmlbase.baseUri, xmlbase.depth)
				}
			}
		}
	}
}
