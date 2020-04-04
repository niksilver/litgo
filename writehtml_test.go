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
		// Added to in section 2 (but we won't check this output file)
		"ch1/first.md": "# Section two\n" +
			"``` Chunk one\n" +
			"chunkonepone(1.1)\n" +
			"```\n" +
			// Added to in section 1
			// Used in sections 2.1 and 3
			"## Section two p one\n" +
			"``` Chunk two\n" +
			"@{Chunk one}\n" +
			"```\n" +
			"# Section three\n" +
			"``` Chunk three\n" +
			"@{Chunk one}\n" +
			"````\n",
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
		"Used in":     `first.html#section-2.1"`,
		"Used in se":  `first.html#section-3"`,
	}

	for seek, exp := range expected {
		found := false
		for _, line := range outStrs {
			if strings.Contains(line, seek) {
				found = true
				if !strings.Contains(line, exp) {
					t.Errorf("Found '%s', in '%s' but it doesn't contain '%s'", seek, line, exp)
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
