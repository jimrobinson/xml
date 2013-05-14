package xmlbase

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
)

type XmlBaseTestUri struct {
	attr xml.Name
	iri  string
}

type XmlBaseTest struct {
	example string
	resolve []XmlBaseTestUri
}

var xmlBaseTests = []XmlBaseTest{
	XmlBaseTest{
		example: `
			<doc
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
			</doc>`,
		resolve: []XmlBaseTestUri{
			{attr: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, iri: "http://example.org/today/new.xml"},
			{attr: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, iri: "http://example.org/hotpicks/pick1.xml"},
			{attr: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, iri: "http://example.org/hotpicks/pick2.xml"},
			{attr: xml.Name{Space: "http://www.w3.org/1999/xlink", Local: "href"}, iri: "http://example.org/hotpicks/pick3.xml"},
		},
	},
	XmlBaseTest{
		example: `<e1 xml:base="http://example.org/wine/"><e2 xml:base="rosé"/></e1>`,
		resolve: []XmlBaseTestUri{
			{attr: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, iri: "http://example.org/wine/"},
			{attr: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, iri: "http://example.org/wine/rosé"},
		},
	},
	XmlBaseTest{
		example: `<elt xml:base="http://www.example.org/~Dürst/"/>`,
		resolve: []XmlBaseTestUri{
			{attr: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, iri: "http://www.example.org/~Dürst/"},
		},
	},
	XmlBaseTest{
		example: `<outer xml:base="http://www.example.org/one/two"> <inner xml:base=""/> </outer>`,
		resolve: []XmlBaseTestUri{
			{attr: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, iri: "http://www.example.org/one/two"},
			{attr: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, iri: "http://www.example.org/one/two"},
		},
	},
	XmlBaseTest{
		example: `<outer xml:base="http://www.example.org/one/two"> <inner xml:base="#frag"/> </outer>`,
		resolve: []XmlBaseTestUri{
			{attr: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, iri: "http://www.example.org/one/two"},
			{attr: xml.Name{Space: xmlBaseSpace, Local: xmlBaseLocal}, iri: "http://www.example.org/one/two"},
		},
	},
}

const verbose = false

func TestXmlBase(t *testing.T) {
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
					if attr.Name.Space == v.resolve[r].attr.Space && attr.Name.Local == v.resolve[r].attr.Local {
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
