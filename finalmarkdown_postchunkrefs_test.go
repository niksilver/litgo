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
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```", // Line 8
		"# T2",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"```", // Line 14
		// Plus one line after the final \n // Line 15
	}
	expected := map[int]string{
		8:  "```",
		9:  "# 2 T2",
		14: "```",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 15 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			15, len(out), b.String())
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
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 9
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 18
		// Post-chunk ref
		// Post-chunk blank
		// plus another when the processor adds a final \n // Line 21
	}
	expected := map[int]string{
		9:  "",
		10: "Added to in section [2](once.html#section-2).",
		11: "",
		18: "",
		19: "Added to in section [1](once.html#section-1).",
		20: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
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

func TestFinalMarkdown_ChunkRefs_AddedToTwice(t *testing.T) {
	s := newState()
	s.setFirstInName("twice.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 9
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 18
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 27
		// Post-chunk ref
		// Post-chunk blank
		// Spare line after final \n // Line 30
	}
	expected := map[int]string{
		9:  "",
		10: "Added to in sections [2](twice.html#section-2) and [2](twice.html#section-2).",
		11: "",

		18: "",
		19: "Added to in sections [1](twice.html#section-1) and [2](twice.html#section-2).",
		20: "",

		27: "",
		28: "Added to in sections [1](twice.html#section-1) and [2](twice.html#section-2).",
		29: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 30 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			30, len(out), b.String())
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
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 9
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 18
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 27
		// Post-chunk ref
		// Post-chunk blank
		"# Title 3",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```",
		// Post-chunk blank // Line 36
		// Post-chunk ref
		// Post-chunk blank
		// Spare line after final \n // Line 39
	}
	expected := map[int]string{
		9:  "",
		10: "Added to in sections [2](thrice.html#section-2), [2](thrice.html#section-2) and [3](thrice.html#section-3).",
		11: "",

		18: "",
		19: "Added to in sections [1](thrice.html#section-1), [2](thrice.html#section-2) and [3](thrice.html#section-3).",
		20: "",

		27: "",
		28: "Added to in sections [1](thrice.html#section-1), [2](thrice.html#section-2) and [3](thrice.html#section-3).",
		29: "",

		36: "",
		37: "Added to in sections [1](thrice.html#section-1), [2](thrice.html#section-2) and [2](thrice.html#section-2).",
		38: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 39 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			39, len(out), b.String())
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
			"Added to in section [1.1](ch1/first.html#section-1.1)",
			"Used in section [2](ch2/second.html#section-2)",
		},
		"ch1/first.md": {
			"Added to in section [1](../book.html#section-1)",
			"Used in section [2](../ch2/second.html#section-2)",
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
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"```", // Line 7
		// Plus one line after the final \n // Line 8
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 8 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			8, len(out), b.String())
	}
}

func TestFinalMarkdown_ChunkRefs_UsedOnce(t *testing.T) {
	s := newState()
	s.setFirstInName("once.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"Chunk content",
		"```",
		// Post-chunk blank // Line 9
		// Post-chunk ref
		// Post-chunk blank
		"# T2",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"@{Chunk one}",
		"```", // Not added to, so no post-chunk refs
		// plus another when the processor adds a final \n // Line 19
	}
	expected := map[int]string{
		9:  "",
		10: "Used in section [2](once.html#section-2).",
		11: "",
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

func TestFinalMarkdown_ChunkRefs_UsedTwice(t *testing.T) {
	s := newState()
	s.setFirstInName("twice.md")
	d := newDoc()
	lines := []string{
		"# Title", // Line 1
		"",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"  # Some comment",
		"  @{Chunk two}",
		"  More chunk content",
		"```",
		"# T2", // Line 11
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"  Some content here",
		"```",
		// Post-chunk blank // Line 18
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk three",
		"@{Chunk two}",
		"```",
		// Spare line after final \n // Line 28
	}
	expected := map[int]string{
		18: "",
		19: "Used in sections [1](twice.html#section-1) and [2](twice.html#section-2).",
		20: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 28 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			28, len(out), b.String())
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
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk one",
		"  @{Chunk two}",
		"```",
		"# T2", // Line 9
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk two",
		"```",
		// Post-chunk blank // Line 15
		// Post-chunk ref
		// Post-chunk blank
		"",
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk three",
		"  Some code",
		"  @{Chunk two}",
		"```",
		"# Title 3", // Line 26
		// Styling for chunk name
		// Chunk name header
		// Blank line
		"``` Chunk four",
		"@{Chunk two}",
		"```",
		// Spare line after final \n // Line 33
	}
	expected := map[int]string{
		15: "",
		16: "Used in sections [1](thrice.html#section-1), [2](thrice.html#section-2) and [3](thrice.html#section-3).",
		17: "",
	}
	r := strings.NewReader(strings.Join(lines, "\n"))

	processContent(r, &s, &d)
	d.lat = compileLattice(d.chunks)
	b := finalMarkdown(s.inName, &d)
	out := strings.Split(b.String(), "\n")

	if len(out) != 33 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			33, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
}
