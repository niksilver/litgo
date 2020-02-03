package main

import (
	"strings"
	"testing"
)

func TestFinalMarkdown_ChunkStart_ChunkNamesGiven(t *testing.T) {
	d := newDoc()
	s := newState()
	s.setFirstInName("test.md")
	lines := []string{
		"# Section one", // Line 1
		"",
		// Styling for chunk name  // Line 3
		// Chunk name header  // Line 4
		// Blank link after chunk name  // Line 5
		"``` Chunk one",
		"Content 1.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		"",
		// Styling for chunk name  // Line 13
		// Chunk name header  // Line 14
		// Blank link after chunk name  // Line 15
		"``` Chunk two",
		"Content 2",
		"```",
		"",
		// Styling for chunk name  // Line 20
		// Chunk name header  // Line 21
		// Blank link after chunk name  // Line 22
		"``` Chunk one",
		"Content 1.2",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		3:  "{.chunk-name}",
		4:  "<a name=\"Chunk-one\"></a>Chunk one",
		5:  "",
		13: "{.chunk-name}",
		14: "<a name=\"Chunk-two\"></a>Chunk two",
		15: "",
		20: "{.chunk-name}",
		21: "Chunk one",
		22: "",
	}
	content := strings.NewReader(strings.Join(lines, "\n"))

	processContent(content, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}
