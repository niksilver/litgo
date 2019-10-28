package main

import "testing"

func TestProcessContent(t *testing.T) {
	// Make sure proc is called at least once normally
	called := false
	proc := func(in string) {
		called = true
	}
	processContent([]byte("Hello"), proc)
	if !called {
		t.Error("proc should have been called at least once")
	}

	// Process three lines in order
	lines := make([]string, 0)
	proc = func(in string) {
		lines = append(lines, in)
	}
	processContent([]byte("One\nTwo\nThree"), proc)
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
