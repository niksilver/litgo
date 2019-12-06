package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestProcessContent(t *testing.T) {

	// Make sure proc is called at least once normally
	called := false
	s := newState()
	mockProc := func(s *state, in string) { called = true }

	processContent([]byte("Hello"), &s, mockProc)

	if !called {
		t.Error("proc should have been called at least once")
	}

	// Process three lines in order
	lines := make([]string, 0)
	s = newState()
	mockProc = func(s *state, in string) { lines = append(lines, in) }

	processContent([]byte("One\nTwo\nThree"), &s, mockProc)

	if len(lines) != 3 {
		t.Errorf("Should have returned 3 lines but got %d", len(lines))
	}
	expected := []string{"One", "Two", "Three"}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line[%d] should be %q but was %q", i, exp, lines[i])
		}
	}
}

func TestProcForMarkdown(t *testing.T) {
	s := newState()
	cs := []struct {
		line     string // Next line
		markdown string // Accumulated markdown
	}{
		{"One",
			"One\n"},
		{"```",
			"One\n```\n"},
		{"Code 1",
			"One\n```\nCode 1\n"},
		{"Code 2",
			"One\n```\nCode 1\nCode 2\n"},
		{"```",
			"One\n```\nCode 1\nCode 2\n```\n"},
		{"End",
			"One\n```\nCode 1\nCode 2\n```\nEnd\n"},
	}

	for i, c := range cs {
		proc(&s, c.line)
		if s.markdown.String() != c.markdown {
			t.Errorf("Line %d: Expected markdown %q but got %q",
				i+1, c.markdown, s.markdown.String())
		}
	}
}

func TestProcForInChunks(t *testing.T) {
	s := newState()
	cs := []struct {
		line    string // Next line
		inChunk bool   // Expected values...
	}{
		{"One", false},
		{"```", true},
		{"Code 1", true},
		{"Code 2", true},
		{"```", false},
		{"End", false},
	}

	for i, c := range cs {
		proc(&s, c.line)
		if s.inChunk != c.inChunk {
			t.Errorf("Line %d: Expected inChunk=%v but got %v",
				i+1, c.inChunk, s.inChunk)
		}
	}
}

func TestProcForChunkNames(t *testing.T) {
	s := newState()
	lines := []string{
		"``` First",
		"Code line 1",
		"Code line 2",
		"```",
		"",
		"``` Second",
		"Code line 3",
		"```",
		"The end",
	}
	first := []string{"Code line 1", "Code line 2"}
	second := []string{"Code line 3"}

	for _, line := range lines {
		proc(&s, line)
	}
	actFirst := s.chunks["First"].code
	if !reflect.DeepEqual(actFirst, first) {
		t.Errorf("Chunk First should be %#v but got %#v",
			first, actFirst)
	}
	actSecond := s.chunks["Second"].code
	if !reflect.DeepEqual(actSecond, second) {
		t.Errorf("Chunk Second should be %#v but got %#v",
			first, actSecond)
	}
}

func TestProcForChunkDetails(t *testing.T) {
	s := newState()
	lines := []string{
		"``` First",
		"Code line 1",
		"Code line 2",
		"```",
		"# Heading",
		"``` Second",
		"Code line 3",
		"```",
		"",
		"``` First", // Appending to a chunk
		"Code line 4",
		"```",
		"The end",
	}
	sec0 := section{[]int(nil), ""}
	sec1 := section{[]int{1}, "Heading"}
	expected := map[string]chunk{
		"First": chunk{
			[]int{1, 10},
			[]section{sec0, sec1},
			[]string{"Code line 1", "Code line 2", "Code line 4"},
			[]int{2, 3, 11},
		},
		"Second": chunk{
			[]int{6},
			[]section{sec1},
			[]string{"Code line 3"},
			[]int{7},
		},
	}

	for _, line := range lines {
		proc(&s, line)
	}

	if len(s.chunks) != 2 {
		t.Errorf("Expected 2 chunks but got %d", len(s.chunks))
	}
	if !reflect.DeepEqual(expected["First"], *s.chunks["First"]) {
		t.Errorf("Expected First chunk to be\n%#v\nbut got\n%#v",
			expected["First"], *s.chunks["First"])
	}
	if !reflect.DeepEqual(expected["Second"], *s.chunks["Second"]) {
		t.Errorf("Expected Second chunk to be\n%#v\nbut got\n%#v",
			expected["Second"], *s.chunks["Second"])
	}
}

func TestProcForWarningsAroundChunks(t *testing.T) {
	s := newState()
	s.fname = "testfile.lit"
	lines := []string{
		"Title",
		"",
		"``` Okay chunk",
		"Chunk content",
		"```",
		"",
		"```", // Chunk start without name
		"```",
		"",
		"``` Another chunk",
		"Chunk content", // Chunk does not end
	}
	content := []byte(strings.Join(lines, "\n"))
	expected := []struct {
		fname string
		line  int
		subs  string
	}{
		{"testfile.lit", 7, "no name"},
		{"testfile.lit", 11, "chunk not closed"},
	}

	processContent(content, &s, proc)

	nWarn := len(s.warnings)
	if nWarn != len(expected) {
		t.Errorf("Expected %d warnings, but got %d", len(expected), nWarn)
	}
	for i, w := range expected {
		if i+1 > nWarn {
			t.Errorf("Warning index %d missing, expected %v", i, w)
			continue
		}
		if expected[i].line != s.warnings[i].line ||
			expected[i].fname != s.warnings[i].fname ||
			!strings.Contains(s.warnings[i].msg, expected[i].subs) {
			t.Errorf("Expected warning index %d to be %v but got %v",
				i, w, s.warnings)
		}
	}
}

func TestProcForChunkRefs(t *testing.T) {
	s := newState()
	s.fname = "testfile.lit"
	lines := []string{
		"Opening text", // Line 1
		"",
		"``` Chunk one",
		"Chunk content",
		"```", // Line 5
		"# First section",
		"``` Chunk two",
		"# Comment, not section heading",
		"```", // Line 9
		"",
		"``` Chunk three",
		"More chunk content",
		"```", // Line 13
	}
	sec0 := section{[]int(nil), ""}
	sec1 := section{[]int{1}, "First section"}
	content := []byte(strings.Join(lines, "\n"))
	expected := map[int]chunkRef{
		5:  chunkRef{"Chunk one", sec0},
		9:  chunkRef{"Chunk two", sec1},
		13: chunkRef{"Chunk three", sec1},
	}

	processContent(content, &s, proc)

	if len(s.chunkRefs) != len(expected) {
		t.Errorf("Expected %d chunk refs but got %d. Map is %#v",
			len(expected), len(s.chunkRefs), s.chunkRefs)
		return
	}
	for lNum, ref := range expected {
		if !reflect.DeepEqual(s.chunkRefs[lNum], ref) {
			t.Errorf("For line %d expected chunk %#v but got %#v",
				lNum, ref, s.chunkRefs[lNum])
		}
	}
}

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
	content := []byte(strings.Join(lines, "\n"))

	processContent(content, &s, proc)
	b := markdownWithChunkRefs(&s)
	out := strings.Split(b.String(), "\n")

	if len(out) != 9 {
		t.Errorf("Expected %d lines but got %d:\n%q",
			9, len(out), b.String())
	}
	for n, s := range expected {
		if out[n-1] != s {
			t.Errorf("Expected line %d to be %q but got %q",
				n, s, out[n-1])
		}
	}
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
	content := []byte(strings.Join(lines, "\n"))

	processContent(content, &s, proc)
	b := markdownWithChunkRefs(&s)
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
	content := []byte(strings.Join(lines, "\n"))

	processContent(content, &s, proc)
	b := markdownWithChunkRefs(&s)
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
	content := []byte(strings.Join(lines, "\n"))

	processContent(content, &s, proc)
	b := markdownWithChunkRefs(&s)
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
