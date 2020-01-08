package main

import (
	"io"
	"strings"
	"testing"
)

func TestFinalMarkdown_AnchorsAndNumbers_AtSectionStarts(t *testing.T) {
	d := newDoc()
	s := newState()
	s.setFirstInName("test.md")
	lines := []string{
		"# Section one", // Line 1
		"Content 1a",
		"## Section one point one", // Line 3
		"Content 1.1a",
		"Content 1.1b",
		"# Section two", // Line 6
		"Content 2a",
	}
	expected := map[int]string{
		1: "# <a name=\"section-1\"></a>1 Section one",
		3: "## <a name=\"section-1.1\"></a>1.1 Section one point one",
		6: "# <a name=\"section-2\"></a>2 Section two",
	}
	content := strings.NewReader(strings.Join(lines, "\n"))

	processContent(content, &s, &d)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestFinalMarkdown_Anchors_AtFileStarts(t *testing.T) {
	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "First line 1\nFirst line 2\n# Section One",
		"second.md": "Second line 1\nSecond line 2",
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	expected := map[string][]struct {
		lNum int
		cont string
	}{
		"book.md": {
			{1, "<a name=\"section-0\"></a>"},
			{2, "* [First chapter](first.md)"},
		},
		"first.md": {
			{1, "<a name=\"section-0\"></a>"},
			{2, "First line 1"},
		},
		"second.md": {
			{1, "<a name=\"section-1\"></a>"},
			{2, "Second line 1"},
		},
	}

	firstPassForAll(&s, &d)

	for fName, slc := range expected {
		b := finalMarkdown(fName, &d)
		out := strings.Split(b.String(), "\n")

		for _, exp := range slc {
			lNum, cont := exp.lNum, exp.cont
			if out[lNum-1] != cont {
				t.Errorf("Expected %s:%d to be %q but got %q",
					fName, lNum, cont, out[lNum-1])
			}
		}

	}
}
