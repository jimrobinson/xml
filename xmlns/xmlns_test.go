package xmlns

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
)

var _ = fmt.Errorf

type XmlNamespaceTest struct {
	sample string
	prefix []Prefix
	uri    []Uri
}

var nsSamples = []XmlNamespaceTest{
	XmlNamespaceTest{
		`<a><b></b></a>`,
		[]Prefix{
			Prefix{},
			Prefix{},
		},
		[]Uri{
			Uri{},
			Uri{},
		},
	},
	XmlNamespaceTest{
		`<a:a
			xmlns:a="ns-a"
			xmlns:b="ns-b"
			xmlns:c="ns-b">
			<b:b
				xmlns:d="ns-d"
				xmlns:e="ns-e">
				<b:b><b:b>
					<c:c
						xmlns:c="ns-c"
						xmlns:a="ns-1">
						<d:d/>
					</c:c>
				</b:b></b:b>
			</b:b>
		</a:a>`,
		[]Prefix{
			Prefix{"a": "ns-a", "b": "ns-b", "c": "ns-b"},
			Prefix{"a": "ns-a", "b": "ns-b", "c": "ns-b", "d": "ns-d", "e": "ns-e"},
			Prefix{"a": "ns-a", "b": "ns-b", "c": "ns-b", "d": "ns-d", "e": "ns-e"},
			Prefix{"a": "ns-a", "b": "ns-b", "c": "ns-b", "d": "ns-d", "e": "ns-e"},
			Prefix{"a": "ns-1", "b": "ns-b", "c": "ns-c", "d": "ns-d", "e": "ns-e"},
			Prefix{"a": "ns-1", "b": "ns-b", "c": "ns-c", "d": "ns-d", "e": "ns-e"},
		},
		[]Uri{
			Uri{"ns-a": []string{"a"}, "ns-b": []string{"b", "c"}},
			Uri{"ns-a": []string{"a"}, "ns-b": []string{"b", "c"}, "ns-d": []string{"d"}, "ns-e": []string{"e"}},
			Uri{"ns-a": []string{"a"}, "ns-b": []string{"b", "c"}, "ns-d": []string{"d"}, "ns-e": []string{"e"}},
			Uri{"ns-a": []string{"a"}, "ns-b": []string{"b", "c"}, "ns-d": []string{"d"}, "ns-e": []string{"e"}},
			Uri{"ns-1": []string{"a"}, "ns-b": []string{"b"}, "ns-c": []string{"c"}, "ns-d": []string{"d"}, "ns-e": []string{"e"}},
			Uri{"ns-1": []string{"a"}, "ns-b": []string{"b"}, "ns-c": []string{"c"}, "ns-d": []string{"d"}, "ns-e": []string{"e"}},
		},
	},
}

func TestPush(t *testing.T) {
	xmlns := NewXmlNamespace()

	for i := range nsSamples {
		j := 0
		dec := xml.NewDecoder(strings.NewReader(nsSamples[i].sample))
		for {
			tok, err := dec.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Error("nsSamples:", i, err)
			}
			switch node := tok.(type) {
			case xml.StartElement:
				xmlns.Push(node)
				checkState("push", j, xmlns, nsSamples[i].prefix[j], nsSamples[i].uri[j], t)
				j++
			case xml.EndElement:
				j--
				checkState("pop", j, xmlns, nsSamples[i].prefix[j], nsSamples[i].uri[j], t)
				xmlns.Pop()
			}
		}
	}
}

func TestCheck(t *testing.T) {
	sample := []string{`<a a:b="c" xmlns:a="b"/>`, `<a a:b="c"/>`}
	for i := 0; i < len(sample); i++ {
		dec := xml.NewDecoder(strings.NewReader(sample[i]))
		xmlns := NewXmlNamespace()
		for {
			tok, err := dec.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Error(err)
			}
			switch node := tok.(type) {
			case xml.StartElement:
				xmlns.Push(node)
				err = xmlns.Check(node)
				if i == 0 && err != nil {
					t.Error("unexpected error returned by xmlns.Check:", err)
				}
				if i == 1 && err == nil {
					t.Error("expected an error to be returned on xmlns.Check")
				}
			case xml.EndElement:
				xmlns.Pop()
			}
		}
	}

}

func checkState(s string, n int, xmlns *XmlNamespace, prefix Prefix, uri Uri, t *testing.T) {
	realPrefix := xmlns.InScope().Prefix
	if len(prefix) != len(realPrefix) {
		t.Errorf("failed test %s.%d: expected %d namespaces, got %d: expected %v, got %v",
			s, n, len(prefix), len(realPrefix), prefix, realPrefix)
	}
	for k, v := range prefix {
		if realPrefix[k] != v {
			t.Errorf("failed test %s.%d: wanted xmlns:%s='%s', got xmlns:%s='%s'",
				s, n, k, v, k, realPrefix[k])
		}
	}

	for u, p := range uri {
		x := xmlns.Prefix(u)
		if p[0] != x {
			t.Errorf("failed test %s.%d: expected xmlns:%s=%s, got xmlns:%s=%s", s, n, p, u, x, u)
		}
	}
}

func BenchmarkParseBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := strings.NewReader(feedXml)
		dec := xml.NewDecoder(r)
		b.StartTimer()
		for {
			tok, err := dec.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
			switch node := tok.(type) {
			case xml.StartElement:
				if node.Name.Space != node.Name.Space {
				}
			case xml.EndElement:
				if node.Name.Space != node.Name.Space {
				}
			}
		}
	}
}

func BenchmarkParsePush(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := strings.NewReader(feedXml)
		dec := xml.NewDecoder(r)
		b.StartTimer()

		nspace := NewXmlNamespace()
		for {
			tok, err := dec.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
			switch node := tok.(type) {
			case xml.StartElement:
				nspace.Push(node)
			case xml.EndElement:
				nspace.Pop()
			}
		}
	}
}

func BenchmarkParseCheck(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := strings.NewReader(feedXml)
		dec := xml.NewDecoder(r)
		b.StartTimer()

		nspace := NewXmlNamespace()
		for {
			tok, err := dec.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
			switch node := tok.(type) {
			case xml.StartElement:
				nspace.Push(node)
				if err := nspace.Check(node); err != nil {
					b.Error(err)
				}
			case xml.EndElement:
				nspace.Pop()
			}
		}
	}
}

var feedXml = `<?xml version="1.0"?>
<atom:feed
  xmlns:atom="http://www.w3.org/2005/Atom"
  xmlns:app="http://www.w3.org/2007/app"
  xmlns:c="http://schema.example.org/Compound"
  xmlns:l="http://schema.example.org/Linking"
  xmlns:r="http://schema.example.org/Revision"
  xmlns:hwp="http://schema.example.org/Journal"
  xmlns:nlm="http://schema.example.org/NLM/Journal"
  xmlns:x="http://www.w3.org/1999/xhtml"
  xmlns:openSearch="http://a9.com/-/spec/opensearch/1.1/">
  <atom:title>Search Results</atom:title>
  <atom:link
    rel="self"
    type="application/atom+xml; type=feed"
    href="http://test.example.org/root.atom?query-form=search&amp;with-descendants=yes&amp;atom:link@:href;eq=%2Fabcd%2F1%2F1%2F1.full.pdf"/>
  <atom:updated>2013-06-03T14:21:40.014149-07:00</atom:updated>
  <atom:id>http://test.example.org/root.atom?query-form=search&amp;with-descendants=yes&amp;atom:link@:href;eq=%2Fabcd%2F1%2F1%2F1.full.pdf</atom:id>
  <openSearch:totalResults>1</openSearch:totalResults>
  <openSearch:startIndex>1</openSearch:startIndex>
  <openSearch:itemsPerPage>100</openSearch:itemsPerPage>
  <atom:entry
    nlm:article-type="research-article"
    xml:lang="en-us">
    <r:released>2010-01-15T00:00:00-08:00</r:released>
    <nlm:pub-date
      pub-type="epub-version"
      hwp:start="2010-01-15T00:00:00-08:00">
      <nlm:day>15</nlm:day>
      <nlm:month>1</nlm:month>
      <nlm:year>2010</nlm:year>
    </nlm:pub-date>
    <atom:link
      rel="http://schema.example.org/Compound#parent"
      href="/abcd/1/1.atom"
      c:role="http://schema.example.org/Journal/Issue"/>
    <atom:category
      scheme="http://schema.example.org/Publishing#role"
      term="http://schema.example.org/Journal/Article"/>
    <atom:id>tag:abcd@example.org,2010-01-01:1/1/1</atom:id>
    <atom:title>Recent Progress in the Theories of Things by People</atom:title>
    <atom:author
    nlm:contrib-type="author">
      <atom:name>John Doe</atom:name>
      <nlm:name name-style="western" hwp:sortable="Doe John">
        <nlm:surname>Doe</nlm:surname>
        <nlm:given-names>John</nlm:given-names>
      </nlm:name>
    </atom:author>
    <atom:published>2010-01-01T00:00:00Z</atom:published>
    <atom:updated>2011-02-27T08:05:31Z</atom:updated>
    <app:edited>2013-02-02T08:09:19.681899-08:00</app:edited>
    <r:created>2010-01-01T00:00:00-08:00</r:created>
    <r:received>2010-01-01T00:00:00-08:00</r:received>
    <nlm:title-group>
      <nlm:article-title hwp:id="article-title-1">Recent Progress in the Theories of Things by People</nlm:article-title>
    </nlm:title-group>
    <nlm:journal-meta>
      <nlm:journal-id journal-id-type="hwp">abcd</nlm:journal-id>
      <nlm:journal-id journal-id-type="nlm-ta">Random Publisher ABCD</nlm:journal-id>
      <nlm:journal-title>Journal of ABCD</nlm:journal-title>
      <nlm:abbrev-journal-title abbrev-type="publisher">ABCD</nlm:abbrev-journal-title>
      <nlm:issn pub-type="ppub">0027-8424</nlm:issn>
      <nlm:issn pub-type="epub">1091-6490</nlm:issn>
      <nlm:publisher>
        <nlm:publisher-name>Random Publisher ABCD</nlm:publisher-name>
      </nlm:publisher>
    </nlm:journal-meta>
    <nlm:volume-id
    pub-id-type="other"
    hwp:sub-type="slug">1</nlm:volume-id>
    <nlm:volume-id pub-id-type="other" hwp:sub-type="tag">1</nlm:volume-id>
    <nlm:issue-id
      pub-id-type="other"
      hwp:sub-type="pisa">abcd;1/1</nlm:issue-id>
    <nlm:issue-id
      pub-id-type="other"
      hwp:sub-type="slug">1</nlm:issue-id>
    <nlm:issue-id
      pub-id-type="other"
      hwp:sub-type="tag">1/1</nlm:issue-id>
    <nlm:article-id
      pub-id-type="pmid">1652094940911</nlm:article-id>
    <nlm:article-id
      pub-id-type="other"
      hwp:sub-type="pisa">abcd;1/1/1</nlm:article-id>
    <nlm:article-id
      pub-id-type="other"
      hwp:sub-type="slug">1</nlm:article-id>
    <nlm:article-id
      pub-id-type="other"
      hwp:sub-type="tag">1/1/1</nlm:article-id>
    <nlm:pub-id
      pub-id-type="pmid">1652094940911</nlm:pub-id>
    <nlm:pub-id
      pub-id-type="other"
      hwp:sub-type="pisa">abcd;1/1/1</nlm:pub-id>
    <nlm:pub-id
      pub-id-type="other"
      hwp:sub-type="slug">1</nlm:pub-id>
    <nlm:pub-id
      pub-id-type="other"
      hwp:sub-type="tag">1/1/1</nlm:pub-id>
    <nlm:article-categories>
      <nlm:subj-group subj-group-type="heading">
        <nlm:subject>Mathematics</nlm:subject>
      </nlm:subj-group>
    </nlm:article-categories>
    <nlm:pub-date
      pub-type="ppub"
      hwp:start="2010-01">
      <nlm:month>01</nlm:month>
      <nlm:year>2010</nlm:year>
    </nlm:pub-date>
    <nlm:pub-date
      pub-type="hwp-created"
      hwp:start="2010-01-01T00:00:00-08:00">
      <nlm:day>1</nlm:day>
      <nlm:month>1</nlm:month>
      <nlm:year>2010</nlm:year>
    </nlm:pub-date>
    <nlm:pub-date
    pub-type="hwp-received"
      hwp:start="2010-01-01T00:00:00-08:00">
      <nlm:day>1</nlm:day>
      <nlm:month>1</nlm:month>
      <nlm:year>2010</nlm:year>
    </nlm:pub-date>
    <nlm:pub-date
    pub-type="epub"
      hwp:start="2010-01-15T00:00:00-08:00">
      <nlm:day>15</nlm:day>
      <nlm:month>1</nlm:month>
      <nlm:year>2010</nlm:year>
    </nlm:pub-date>
    <nlm:volume>1</nlm:volume>
    <nlm:issue>1</nlm:issue>
    <nlm:fpage>1</nlm:fpage>
    <nlm:lpage>4</nlm:lpage>
    <atom:link rel="self"
    href="/abcd/1/1/1.atom"/>
    <atom:link rel="edit"
    href="/abcd/1/1/1.atom"/>
    <atom:link
      href="/abcd/1/1/1"
      c:role="http://schema.example.org/Publishing/builtin"/>
    <atom:link
      rel="alternate"
      href="/abcd/1/1/1.atom?form=feed"
      c:role="http://schema.example.org/Publishing/builtin"
      type="application/atom+xml; type=feed"/>
    <atom:link
      rel="http://schema.example.org/Publishing#model"
      href="/abcd/1/1/1.model"/>
    <atom:link
      rel="alternate"
      href="/abcd/1/1/1.full.pdf"
      c:role="http://schema.example.org/alternate/full-text"
      type="application/pdf"
      hreflang="en-us" />
    <atom:link
      rel="http://schema.example.org/Publishing#edit-alternate"
      href="/abcd/1/1/1.full.pdf"
      c:role="http://schema.example.org/alternate/full-text"
      type="application/pdf"
      hreflang="en-us" />
    <atom:link
      rel="alternate"
      href="forthcoming:yes"
      c:role="http://schema.example.org/alternate/original"
      type="application/xml"/>
    <atom:link
      rel="alternate"
      href="/abcd/1/1/1.source.xml"
      c:role="http://schema.example.org/alternate/source"
      type="application/xml"/>
    <atom:link
      rel="http://schema.example.org/Publishing#edit-alternate"
      href="/abcd/1/1/1.source.xml"
      c:role="http://schema.example.org/alternate/source"
      type="application/xml"/>
    <atom:link
      rel="alternate"
      href="/abcd/1/1/1.concepts.rdf"
      c:role="http://schema.example.org/alternate/concepts"
      type="application/rdf+xml"/>
    <atom:link
      rel="http://schema.example.org/Publishing#edit-alternate"
      href="/abcd/1/1/1.concepts.rdf"
      c:role="http://schema.example.org/alternate/concepts"
      type="application/rdf+xml"/>
  </atom:entry>
</atom:feed>`
