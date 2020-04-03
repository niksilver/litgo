package main

import (
	"io"
	"strings"
	"testing"
)

func TestWriteHTML_PostChunkRefsLinkToHTML(t *testing.T) {
	data := map[string]string{
		"book.md": "# Section one\n" +
			"* [Chap 1](ch1/first.md)\n" +
			"``` Chunk one\n" +
			"chunkone(1)\n" +
			"```\n",
		// Added to in section 1.1
		"ch1/first.md": "## Section one p one\n" +
			"``` Chunk one\n" +
			"chunkonepone(1.1)\n" +
			"```\n" +
			// Added to in section 1
			"# Section two\n" +
			"``` Chunk two\n" +
			"@{Chunk one}\n" +
			"````\n",
		// Used in sections 1 and 1.1
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)
	d.lat = compileLattice(d.chunks)
	bDoc := newBuilderDoc(d)
	if err := writeHTML("ch1/first.md", "first.html", &bDoc.doc); err != nil {
		t.Errorf("writeHTML error: %s", err.Error())
	}

	out, okay := bDoc.outputs["first.html"]
	if !okay {
		t.Errorf("No output to Builder out.html")
	}

	outStrs := strings.Split(out.String(), "\n")

	expected := map[string]string{
		"Added to in": `book.html#section-1"`,
		"Used in":     `first.html#section-1"`,
		"Used in se":  `first.html#section-1.1"`,
	}

	for seek, exp := range expected {
		found := false
		for idx, line := range outStrs {
			if strings.Contains(line, seek) {
				found = true
				if !strings.Contains(line, exp) {
					t.Errorf("Found line %d, '%s', with '%s' but it doesn't contain '%s'", idx+1, line, seek, exp)
				}
			}
		}
		if !found {
			t.Errorf("Didn't find a line containing '%s'", seek)
		}
	}

	//	out := strings.Split(b.String(), "\n")
	//
	//	if len(out) != 15 {
	//		t.Errorf("Expected %d lines but got %d:\n%q",
	//			15, len(out), b.String())
	//	}
	//	for n, s := range expected {
	//		if stripHTML(out[n-1]) != s {
	//			t.Errorf("Expected line %d to be %q but got %q",
	//				n, s, out[n-1])
	//		}
	//	}
}
