package main

import (
	"io"
	"regexp"
	"strings"
	"testing"
)

func TestFinalMarkdown_ChunkRefs_AddedToNowhereElse(t *testing.T) {
	s := newState()
	s.setFirstInName("chunktest.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```", // Line 7
		"# T2",
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"```", // Line 12
		// Plus one line after the final \n // Line 13
	}
	expected := map[int]string{
		7:  "```",
		8:  "# 2 T2",
		12: "```",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 13 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			13, len(out), b.String())
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

func TestFinalMarkdown_ChunkRefs_AddedToOnce(t *testing.T) {
	s := newState()
	s.setFirstInName("once.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 8
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 16
		// Post-chunk ref
		// Post-chunk blank
		// plus another when the processor adds a final \n // Line 19
	}
	expected := map[int]string{
		8:  "",
		9:  "Added to in section [2](once.md#section-2).",
		10: "",
		16: "",
		17: "Added to in section [1](once.md#section-1).",
		18: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
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

func TestFinalMarkdown_ChunkRefs_AddedToTwice(t *testing.T) {
	s := newState()
	s.setFirstInName("twice.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 8
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 16
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 24
		// Post-chunk ref
		// Post-chunk blank
		// Spare line after final \n // Line 27
	}
	expected := map[int]string{
		8:  "",
		9:  "Added to in sections [2](twice.md#section-2) and [2](twice.md#section-2).",
		10: "",
		16: "",
		17: "Added to in sections [1](twice.md#section-1) and [2](twice.md#section-2).",
		18: "",

		24: "",
		25: "Added to in sections [1](twice.md#section-1) and [2](twice.md#section-2).",
		26: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
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

func TestFinalMarkdown_ChunkRefs_AddedToThrice(t *testing.T) {
	s := newState()
	s.setFirstInName("thrice.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 8
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 16
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 24
		// Post-chunk ref
		// Post-chunk blank
		"# Title 3",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 32
		// Post-chunk ref
		// Post-chunk blank
		// Spare line after final \n // Line 35
	}
	expected := map[int]string{
		8:  "",
		9:  "Added to in sections [2](thrice.md#section-2), [2](thrice.md#section-2) and [3](thrice.md#section-3).",
		10: "",

		16: "",
		17: "Added to in sections [1](thrice.md#section-1), [2](thrice.md#section-2) and [3](thrice.md#section-3).",
		18: "",

		24: "",
		25: "Added to in sections [1](thrice.md#section-1), [2](thrice.md#section-2) and [3](thrice.md#section-3).",
		26: "",

		32: "",
		33: "Added to in sections [1](thrice.md#section-1), [2](thrice.md#section-2) and [2](thrice.md#section-2).",
		34: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 35 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			35, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestFinalMarkdown_ChunkRefs_CrossFileBoundaries(t *testing.T) {
	data := map[string]string{
		"book.md": "# Section one\n" +
			"* [Chap 1](ch1/first.md)\n" +
			"* [Chap 2](ch2/second.md)\n" +
			"``` Chunk one\n" +
			"chunkone(1)\n" +
			"```\n",
		"ch1/first.md": "## Section one p one\n" +
			"``` Chunk one\n" +
			"chunkonepone(1.1)\n" +
			"```\n",
		"ch2/second.md": "# Section two\n" +
			"``` Chunk two\n" +
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

	expected := map[string][]string{
		"book.md": {
			"Added to in section [1.1](ch1/first.md#section-1.1)",
			"Used in section [2](ch2/second.md#section-2)",
		},
		"ch1/first.md": {
			"Added to in section [1](../book.md#section-1)",
			"Used in section [2](../ch2/second.md#section-2)",
		},
	}

	firstPassForAll(&s, &d)
	d.lat = compileLattice(d.chunks)

	for fName, subs := range expected {
		mdown := finalMarkdown(fName, &d).String()
		for _, sub := range subs {
			if !strings.Contains(mdown, sub) {
				t.Errorf("Expected %s to contain %q but it is\n%s",
					fName, sub, mdown)
			}
		}

	}
}

func TestFinalMarkdown_ChunkRefs_UsedNowhere(t *testing.T) {
	s := newState()
	s.setFirstInName("nowhere.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```", // Line 6
		// Plus one line after the final \n // Line 7
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 7 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			7, len(out), b.String())
	}
}

func TestFinalMarkdown_ChunkRefs_UsedOnce(t *testing.T) {
	s := newState()
	s.setFirstInName("once.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 8
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"@{Chunk one}",
		"```", // Not added to, so no post-chunk refs
		// plus another when the processor adds a final \n // Line 17
	}
	expected := map[int]string{
		8:  "",
		9:  "Used in section [2](once.md#section-2).",
		10: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 17 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			17, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestFinalMarkdown_ChunkRefs_UsedTwice(t *testing.T) {
	s := newState()
	s.setFirstInName("twice.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"  # Some comment",
		"  @{Chunk two}",
		"  More chunk content",
		"```",
		"# T2", // Line 10
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"  Some content here",
		"```",
		// Post-chunk blank // Line 16
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Chunk name header
		// Blank line
		"``` Chunk three",
		"@{Chunk two}",
		"```",
		// Spare line after final \n // Line 25
	}
	expected := map[int]string{
		16: "",
		17: "Used in sections [1](twice.md#section-1) and [2](twice.md#section-2).",
		18: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 25 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			25, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}

func TestFinalMarkdown_ChunkRefs_UsedThrice(t *testing.T) {
	s := newState()
	s.setFirstInName("thrice.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"  @{Chunk two}",
		"```",
		"# T2", // Line 8
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"```",
		// Post-chunk blank // Line 13
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Chunk name header
		// Blank line
		"``` Chunk three",
		"  Some code",
		"  @{Chunk two}",
		"```",
		"# Title 3", // Line 23
		// Chunk name header
		// Blank line
		"``` Chunk four",
		"@{Chunk two}",
		"```",
		// Spare line after final \n // Line 29
	}
	expected := map[int]string{
		13: "",
		14: "Used in sections [1](thrice.md#section-1), [2](thrice.md#section-2) and [3](thrice.md#section-3).",
		15: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 29 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			29, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}
