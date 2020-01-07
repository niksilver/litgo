package main

import (
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
