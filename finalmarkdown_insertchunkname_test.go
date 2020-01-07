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
		// Chunk name  // Line 3
		// Blank link after chunk name  // Line 4
		"``` Chunk one",
		"Content 1.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		"",
		// Chunk name  // Line 12
		// Blank link after chunk name  // Line 13
		"``` Chunk two",
		"Content 2",
		"```",
		"",
		// Chunk name  // Line 18
		// Blank link after chunk name  // Line 19
		"``` Chunk one",
		"Content 1.2",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		3:  "Chunk one",
		4:  "",
		12: "Chunk two",
		13: "",
		18: "Chunk one",
		19: "",
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
