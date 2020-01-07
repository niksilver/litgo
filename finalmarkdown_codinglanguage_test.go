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
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk one", // Line 5
		"@{Chunk 1a}",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk one", // Line 13
		"Content 1.2",
		"```",
		// Post-chunk blank
		// Post-chunk ref (added to in...)
		// Post-chunk blank
		// Chunk name header
		// Blank line after chunk name
		"```Chunk 1a", // Line 21
		"Content 1a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
		"# Language two",
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk.two", // Line 30
		"@{Chunk 2a}",
		"```",
		"",
		// Chunk name header
		// Blank line after chunk name
		"``` Chunk 2a", // Line 36
		"Content 2a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		5:  "```one",
		7:  "```",
		13: "```one",
		15: "```",
		21: "```one",
		30: "```two",
		36: "```two",
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
		// Chunk name header
		// Blank line
		"```", // Missing language // Line 5
		"@{Chunk 1a}",
		"```",
		"",
		// Chunk name header
		// Blank line
		"```Chunk 1a", // Line 11
		"Content 1a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
		"# Language two",
		// Chunk name header
		// Blank line
		"``` Chunk.two", // Line 20
		"@{Chunk 2a}",
		"```",
		"",
		// Chunk name header
		// Blank line
		"``` Chunk 2a", // Line 26
		"Content 2a.1",
		"```",
		// Post-chunk blank
		// Post-chunk ref (used in...)
		// Post-chunk blank
	}
	expected := map[int]string{
		5:  "```",
		7:  "```",
		11: "```",
		20: "```two",
		26: "```two",
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
