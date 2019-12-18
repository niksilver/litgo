package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestProcessContent(t *testing.T) {

	// Make sure proc is called at least once normally
	called := false
	d := newDoc()
	mockProc := func(s *state, d *doc, in string) { called = true }

	processContent(strings.NewReader("Hello"), &state{}, &d, mockProc)

	if !called {
		t.Error("proc should have been called at least once")
	}

	// Process three lines in order
	lines := make([]string, 0)
	d = newDoc()
	mockProc = func(s *state, d *doc, in string) { lines = append(lines, in) }

	processContent(strings.NewReader("One\nTwo\nThree"),
		&state{}, &d, mockProc)

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
	s := state{inName: "markdown.md"}
	d := newDoc()
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
		proc(&s, &d, c.line)
		mdb := d.markdown[s.inName]
		if mdb.String() != c.markdown {
			t.Errorf("Line %d: Expected markdown %q but got %q",
				i+1, c.markdown, mdb.String())
		}
	}
}

func TestProcForInChunks(t *testing.T) {
	d := newDoc()
	s := state{}
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
		proc(&s, &d, c.line)
		if s.inChunk != c.inChunk {
			t.Errorf("Line %d: Expected inChunk=%v but got %v",
				i+1, c.inChunk, s.inChunk)
		}
	}
}

func TestProcForChunkNames(t *testing.T) {
	d := newDoc()
	s := state{}
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
		proc(&s, &d, line)
	}
	actFirst := d.chunks["First"].code
	if !reflect.DeepEqual(actFirst, first) {
		t.Errorf("Chunk First should be %#v but got %#v",
			first, actFirst)
	}
	actSecond := d.chunks["Second"].code
	if !reflect.DeepEqual(actSecond, second) {
		t.Errorf("Chunk Second should be %#v but got %#v",
			first, actSecond)
	}
}

func TestProcForChunkDetails(t *testing.T) {
	d := newDoc()
	s := state{}
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
		proc(&s, &d, line)
	}

	if len(d.chunks) != 2 {
		t.Errorf("Expected 2 chunks but got %d", len(d.chunks))
	}
	if !reflect.DeepEqual(expected["First"], *d.chunks["First"]) {
		t.Errorf("Expected First chunk to be\n%#v\nbut got\n%#v",
			expected["First"], *d.chunks["First"])
	}
	if !reflect.DeepEqual(expected["Second"], *d.chunks["Second"]) {
		t.Errorf("Expected Second chunk to be\n%#v\nbut got\n%#v",
			expected["Second"], *d.chunks["Second"])
	}
}

func TestProcForWarningsAroundChunks(t *testing.T) {
	d := newDoc()
	s := state{inName: "testfile.lit"}
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
	r := strings.NewReader(strings.Join(lines, "\n"))
	expected := []struct {
		fName string
		line  int
		subs  string
	}{
		{"testfile.lit", 7, "no name"},
		{"testfile.lit", 11, "chunk not closed"},
	}

	processContent(r, &s, &d, proc)

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
			expected[i].fName != s.warnings[i].fName ||
			!strings.Contains(s.warnings[i].msg, expected[i].subs) {
			t.Errorf("Expected warning index %d to be %v but got %v",
				i, w, s.warnings)
		}
	}
}

func TestProcForChunkRefs(t *testing.T) {
	d := newDoc()
	s := state{inName: "testfile.lit"}
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
	r := strings.NewReader(strings.Join(lines, "\n"))
	expected := map[int]chunkRef{
		5:  chunkRef{"Chunk one", sec0},
		9:  chunkRef{"Chunk two", sec1},
		13: chunkRef{"Chunk three", sec1},
	}

	processContent(r, &s, &d, proc)

	if len(d.chunkRefs) != len(expected) {
		t.Errorf("Expected %d chunk refs but got %d. Map is %#v",
			len(expected), len(d.chunkRefs), d.chunkRefs)
		return
	}
	for lNum, ref := range expected {
		if !reflect.DeepEqual(d.chunkRefs[lNum], ref) {
			t.Errorf("For line %d expected chunk %#v but got %#v",
				lNum, ref, d.chunkRefs[lNum])
		}
	}
}
