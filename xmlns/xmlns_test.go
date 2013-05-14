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
				t.Fatal("nsSamples:", i, err)
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

func checkState(s string, n int, xmlns *XmlNamespace, prefix Prefix, uri Uri, t *testing.T) {
	realPrefix := xmlns.InScope().Prefix
	if len(prefix) != len(realPrefix) {
		t.Fatalf("failed test %s.%d: expected %d namespaces, got %d: expected %v, got %v",
			s, n, len(prefix), len(realPrefix), prefix, realPrefix)
	}
	for k, v := range prefix {
		if realPrefix[k] != v {
			t.Fatalf("failed test %s.%d: wanted xmlns:%s='%s', got xmlns:%s='%s'",
				s, n, k, v, k, realPrefix[k])
		}
	}

	for u, p := range uri {
		x := xmlns.Prefix(u)
		if p[0] != x {
			t.Fatalf("failed test %s.%d: expected xmlns:%s=%s, got xmlns:%s=%s", s, n, p, u, x, u)
		}
	}
}
