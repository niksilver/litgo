package main

import (
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
	mp1 := mockProc{
		pr: func(in string) { called = true },
	}
	processContent([]byte("Hello"), mp1)
	if !called {
		t.Error("proc should have been called at least once")
	}

	// Process three lines in order
	lines := make([]string, 0)
	mp2 := mockProc{
		pr: func(in string) { lines = append(lines, in) },
	}
	processContent([]byte("One\nTwo\nThree"), mp2)
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
	st := state{}
	st.code = make(map[string]strings.Builder)
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
		st.proc(c.line)
		if st.markdown.String() != c.markdown {
			t.Errorf("Line %d: Expected markdown %q but got %q",
				i+1, c.markdown, st.markdown.String())
		}
	}
}

func TestProcForInChunks(t *testing.T) {
	st := state{}
	st.code = make(map[string]strings.Builder)
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
		st.proc(c.line)
		if st.inChunk != c.inChunk {
			t.Errorf("Line %d: Expected inChunk=%v but got %v",
				i+1, c.inChunk, st.inChunk)
		}
	}
}

func TestProcForChunkNames(t *testing.T) {
	st := state{}
	st.code = make(map[string]strings.Builder)
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
	first := "Code line 1\nCode line 2\n"
	second := "Code line 3\n"

	for _, line := range lines {
		st.proc(line)
	}
	actFirst := st.code["First"]
	if actFirst.String() != first {
		t.Errorf("Chunk First should be %q but got %q",
			first, actFirst.String())
	}
	actSecond := st.code["Second"]
	if actSecond.String() != second {
		t.Errorf("Chunk Second should be %q but got %q",
			first, actSecond.String())
	}
}
