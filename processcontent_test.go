package main

import (
	"strings"
	"testing"
)

func TestProcessContent(t *testing.T) {

	// Make sure proc is called at least once normally
	called := false
	s := newState()
	s.proc = func(s *state, d *doc, in string) { called = true }

	d := newDoc()

	processContent(strings.NewReader("Hello"), &s, &d)

	if !called {
		t.Error("proc should have been called at least once")
	}

	// Process three lines in order
	lines := make([]string, 0)
	s = newState()
	s.proc = func(s *state, d *doc, in string) { lines = append(lines, in) }

	d = newDoc()

	processContent(strings.NewReader("One\nTwo\nThree"), &s, &d)

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
