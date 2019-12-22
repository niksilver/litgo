package main

import (
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
