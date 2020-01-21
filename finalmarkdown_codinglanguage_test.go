package main

import (
	"strings"
	"testing"
)

func TestFinalMarkdown_CodingLanguage_TwoLanguages(t *testing.T) {
	d := newDoc()
	s := newState()
	s.setFirstInName("test.md")
	lines := []string{
		"# Language one", // Line 1
		"",
		// Styling before chunk name
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk one", // Line 6
		"@{Chunk 1a}",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		// Styling before chunk name
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk one", // Line 15
		"Content 1.2",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		// Styling before chunk name
		// Chunk name header
		// Blank line after chunk name
		"```Chunk 1a", // Line 24
		"Content 1a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
		"# Language two",
		// Styling before chunk name
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk.two", // Line 34
		"@{Chunk 2a}",
		"```",
		"",
		// Styling before chunk name
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk 2a", // Line 41
		"Content 2a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		6:  "```one",
		8:  "```",
		15: "```one",
		17: "```",
		24: "```one",
		34: "```two",
		41: "```two",
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

func TestFinalMarkdown_CodingLanguage_MissingLanguages(t *testing.T) {
	d := newDoc()
	s := newState()
	s.setFirstInName("tester.md")
	lines := []string{
		"# Language one", // Line 1
		"",
		// Styling before chunk name
		// Chunk name header
		// Blank line
		"```", // Missing language // Line 6
		"@{Chunk 1a}",
		"```", // Line 8
		"",
		// Styling before chunk name
		// Chunk name header
		// Blank line
		"```Chunk 1a", // Line 13
		"Content 1a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
		"# Language two",
		// Styling before chunk name
		// Chunk name header
		// Blank line
		"``` Chunk.two", // Line 23
		"@{Chunk 2a}",
		"```",
		"",
		// Styling before chunk name
		// Chunk name header
		// Blank line
		"``` Chunk 2a", // Line 30
		"Content 2a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		6:  "```",
		8:  "```",
		13: "```",
		23: "```two",
		30: "```two",
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
