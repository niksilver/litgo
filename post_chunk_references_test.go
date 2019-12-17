package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestProcForMarkdownWithChunkRefs_AddedToNowhereElse(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"Chunk content",
		"```", // Line 5
		"# T2",
		"``` Chunk two",
		"```", // Line 8
		// Plus one line after the final \n // Line 9
	}
	expected := map[int]string{
		5: "```",
		6: "# 2 T2",
		8: "```",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 9 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			9, len(out), b.String())
	}
	for n, s := range expected {
		if stripHTML(out[n-1]) != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func stripHTML(str string) string {
	html, _ := regexp.Compile("<.*>")
	return html.ReplaceAllString(str, "")
}

func TestProcForMarkdownWithChunkRefs_AddedToOnce(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 6
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 12
		// Post-chunk ref
		// Post-chunk blank
		// plus another (15) when the processor adds a final \n
	}
	expected := map[int]string{
		6:  "",
		7:  "Added to in section 2.",
		8:  "",
		12: "",
		13: "Added to in section 1.",
		14: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 15 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			15, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestProcForMarkdownWithChunkRefs_AddedToTwice(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 6
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 12
		// Post-chunk ref
		// Post-chunk blank
		"",
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 18
		// Post-chunk ref
		// Post-chunk blank
		// Spare line after final \n // Line 21
	}
	expected := map[int]string{
		6:  "",
		7:  "Added to in sections 2 and 2.",
		8:  "",
		12: "",
		13: "Added to in sections 1 and 2.",
		14: "",
		18: "",
		19: "Added to in sections 1 and 2.",
		20: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 21 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			21, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestProcForMarkdownWithChunkRefs_AddedToThrice(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 6
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 12
		// Post-chunk ref
		// Post-chunk blank
		"",
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 18
		// Post-chunk ref
		// Post-chunk blank
		"# Title 3",
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 24
		// Post-chunk ref
		// Post-chunk blank
		// Spare line after final \n // Line 27
	}
	expected := map[int]string{
		6:  "",
		7:  "Added to in sections 2, 2 and 3.",
		8:  "",
		12: "",
		13: "Added to in sections 1, 2 and 3.",
		14: "",
		24: "",
		25: "Added to in sections 1, 2 and 2.",
		26: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 27 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			27, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestProcForMarkdownWithChunkRefs_UsedNowhere(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"```", // Line 4
		// Plus one line after the final \n // Line 5
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 5 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			5, len(out), b.String())
	}
}

func TestProcForMarkdownWithChunkRefs_UsedOnce(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 6
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		"``` Chunk two",
		"@{Chunk one}",
		"```", // Not added to, so no post-chunk refs
		// plus another when the processor adds a final \n // Line 13
	}
	expected := map[int]string{
		6: "",
		7: "Used in section 2.",
		8: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 13 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			13, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestProcForMarkdownWithChunkRefs_UsedTwice(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"  # Some comment",
		"  @{Chunk two}",
		"  More chunk content",
		"```",
		"# T2", // Line 8
		"``` Chunk two",
		"  Some content here",
		"```",
		// Post-chunk blank // Line 12
		// Post-chunk ref
		// Post-chunk blank
		"",
		"``` Chunk three",
		"@{Chunk two}",
		"```",
		// Spare line after final \n // Line 19
	}
	expected := map[int]string{
		12: "",
		13: "Used in sections 1 and 2.",
		14: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 19 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			19, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestProcForMarkdownWithChunkRefs_UsedThrice(t *testing.T) {
	s := newState()
	lines := []string{
		"# Title", // Line 1
		"",
		"``` Chunk one",
		"  @{Chunk two}",
		"```",
		"# T2", // Line 6
		"``` Chunk two",
		"```",
		// Post-chunk blank // Line 9
		// Post-chunk ref
		// Post-chunk blank
		"",
		"``` Chunk three",
		"  Some code",
		"  @{Chunk two}",
		"```",
		"# Title 3", // Line 17
		"``` Chunk four",
		"@{Chunk two}",
		"```",
		// Spare line after final \n // Line 21
	}
	expected := map[int]string{
		9:  "",
		10: "Used in sections 1, 2 and 3.",
		11: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, proc)
	s.lat = compileLattice(s.chunks)
	b := finalMarkdown(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 21 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			21, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}
