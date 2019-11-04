package main

import (
	"reflect"
	"strings"
	"testing"
)

type mockProc struct {
	pr func(s string)
}

func (mp mockProc) proc(s string) { mp.pr(s) }

func TestProcessContent(t *testing.T) {

	// Make sure proc is called at least once normally
	called := false
	s := newState()
	s.proc = func(s *state, in string) { called = true }

	processContent([]byte("Hello"), &s)

	if !called {
		t.Error("proc should have been called at least once")
	}

	// Process three lines in order
	lines := make([]string, 0)
	s = newState()
	s.proc = func(s *state, in string) { lines = append(lines, in) }

	processContent([]byte("One\nTwo\nThree"), &s)

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
		s.proc(&s, c.line)
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
		s.proc(&s, c.line)
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
		s.proc(&s, line)
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
		"",
		"``` Second",
		"Code line 3",
		"```",
		"",
		"``` First", // Appending to a chunk
		"Code line 4",
		"```",
		"The end",
	}
	expected := map[string]chunk{
		"First": chunk{
			[]int{1, 10},
			[]string{"Code line 1", "Code line 2", "Code line 4"},
			[]int{2, 3, 11},
		},
		"Second": chunk{
			[]int{6},
			[]string{"Code line 3"},
			[]int{7},
		},
	}

	for _, line := range lines {
		s.proc(&s, line)
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
		line int
		subs string
	}{
		{7, "no name"},
		{11, "chunk not closed"},
	}

	processContent(content, &s)

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
			!strings.Contains(s.warnings[i].msg, expected[i].subs) {
			t.Errorf("Expected warning index %d to be %v but got %v",
				i, w, s.warnings)
		}
	}
}
