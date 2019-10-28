package main

import "testing"

func TestProcessContent(t *testing.T) {
	called := false
	proc := func(in string) {
		called = true
	}
	processContent([]byte("Hello"), proc)
	if !called {
		t.Error("proc should have been called at least once")
	}
}
