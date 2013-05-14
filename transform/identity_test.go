package transform

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

var sampleXml = `<?xml version="1.0" encoding="UTF-8" standalone="no" ?><?xml-stylesheet type="text/xsl" href="/images/Glossary/main.xsl"?><atom:entry xmlns:hw="org.highwire.hpp" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:app="http://www.w3.org/2007/app" xmlns:c="http://schema.highwire.org/Compound" xmlns:l="http://schema.highwire.org/Linking" xmlns:r="http://schema.highwire.org/Revision" xmlns:hwp="http://schema.highwire.org/Journal" xmlns:nlm="http://schema.highwire.org/NLM/Journal" xmlns:x="http://www.w3.org/1999/xhtml">
  <atom:id>http://atom.highwire.org/</atom:id>
  <atom:title>A Test: &amp; &apos; &gt; &lt; &quot;</atom:title>
  <atom:updated>2008-05-02T12:45:11.233099-07:00</atom:updated>
  <r:released r:a="'" r:b='"' r:c="&quot;" r:d='&apos;'>2013-05-01T01:02:03-07:00</r:released>
  <atom:content xml:base="/pnas/109/1/1.full.html" c:role="http://schema.highwire.org/variant/full-text" type="application/xhtml+xml">
    <div xmlns="http://www.w3.org/1999/xhtml" class="article fulltext-view">
      <h1 id="article-title-1">In This Issue</h1>
      <div class="boxed-text" id="boxed-text-1">
        <div id="sec-1" class="subsection">
          <h4>Bacteria might help curb the spread of dengue virus</h4>
          <div id="F1" class="fig pos-float type-figure odd">
            <div class="fig-inline">
              <a href="pending:yes" l:ref-type="journal" hwp:journal="pnas" hwp:volume="109" hwp:issue="1" hwp:article="1" hwp:fragment="F1" l:sub-ref="graphic-1" l:role="expansion" l:type="image/*">
                <img src="pending:yes" l:ref-type="journal" hwp:journal="pnas" hwp:volume="109" hwp:issue="1" hwp:article="1" hwp:fragment="F1" l:sub-ref="graphic-1" l:role="small" l:type="image/*" alt="Figure"/>
              </a>
              <div class="callout">
                <span>View larger version:</span>
                <ul class="callout-links">
                  <li>
                    <a href="pending:yes" l:ref-type="journal" hwp:journal="pnas" hwp:volume="109" hwp:issue="1" hwp:article="1" hwp:fragment="F1" l:role="expansion">In this window</a>
                  </li>
                  <li>
                    <a href="pending:yes" l:ref-type="journal" hwp:journal="pnas" hwp:volume="109" hwp:issue="1" hwp:article="1" hwp:fragment="F1" l:role="expansion" class="in-nw">In a new window</a>
                  </li>
                </ul>
              </div>
            </div>
            <div class="fig-caption">
              <p id="p-1" class="first-child"><em>Aedes albopictus</em> mosquito feeding on human blood. Image courtesy of James Gathany/Centers for Disease Control and Prevention.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </atom:content>
  <!-- <test pattern="SECAM" /><test pattern="NTSC" /> -->
</atom:entry>`

func TestIdentity(t *testing.T) {
	w := new(bytes.Buffer)
	tr, err := NewIdentityTransform(w, "")
	if err != nil {
		t.Fatal(err)
	}

	err = Transform(strings.NewReader(sampleXml), tr)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	x1 := w.String()
	fmt.Println(x1)

	err = compareXml(strings.NewReader(sampleXml), strings.NewReader(x1))
	if err != nil {
		t.Fatal(err)
	}
}

func compareXml(r1, r2 io.Reader) error {
	dec1 := xml.NewDecoder(r1)
	dec2 := xml.NewDecoder(r2)
	for {

		tok1, err1 := dec1.Token()
		tok2, err2 := dec2.Token()
		if err1 != err2 {
			return fmt.Errorf("err1 [%v] != err2 [%v]", err1, err2)
		}
		if err1 == io.EOF {
			break
		}
		if !reflect.DeepEqual(tok1, tok2) {
			return fmt.Errorf("tok1 [%v] != tok2 [%v]", tok1, tok2)
		}
	}
	return nil
}

func BenchmarkIdentity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := new(bytes.Buffer)
		tr, err := NewIdentityTransform(w, "")
		if err != nil {
			b.Fatal(err)
		}
		r := strings.NewReader(sampleXml)
		b.StartTimer()
		err = Transform(r, tr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
