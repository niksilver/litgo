package main

import (
	"strings"
	"testing"
)

func TestFinalMarkDown_TwoCodingLanguages(t *testing.T) {
	s := newState()
	lines := []string{
		"# Language one", // Line 1
		"",
		"``` Chunk one", // Line 3
		"@{Chunk 1a}",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		"``` Chunk one", // Line 9
		"Content 1.2",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		"```Chunk 1a", // Line 15
		"Content 1a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
		"# Language two",
		"``` Chunk.two", // Line 22
		"@{Chunk 2a}",
		"```",
		"",
		"``` Chunk 2a", // Line 26
		"Content 2a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		3:  "```one",
		5:  "```",
		9:  "```one",
		11: "```",
		15: "```one",
		22: "```two",
		26: "```two",
	}
	content := []byte(strings.Join(lines, "\n"))

	processContent(content, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestFinalMarkDown_MissingCodingLanguages(t *testing.T) {
	s := newState()
	lines := []string{
		"# Language one", // Line 1
		"",
		"```", // Missing language // Line 3
		"@{Chunk 1a}",
		"```",
		"",
		"```Chunk 1a", // Line 7
		"Content 1a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
		"# Language two",
		"``` Chunk.two", // Line 14
		"@{Chunk 2a}",
		"```",
		"",
		"``` Chunk 2a", // Line 18
		"Content 2a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		3:  "```",
		5:  "```",
		7:  "```",
		14: "```two",
		18: "```two",
	}
	content := []byte(strings.Join(lines, "\n"))

	processContent(content, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}
