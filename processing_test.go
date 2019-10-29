package main

import "testing"

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

func TestProc(t *testing.T) {
	st := state{}
	cs := []struct {
		line     string // Next line
		markdown string // Accumulated fields...
		inChunk  bool
		code     string
	}{
		{"One",
			"One\n", false, ""},
		{"```",
			"One\n```\n", true, ""},
		{"Code 1",
			"One\n```\nCode 1\n", true, "Code 1\n"},
		{"Code 2",
			"One\n```\nCode 1\nCode 2\n", true, "Code 1\nCode 2\n"},
		{"```",
			"One\n```\nCode 1\nCode 2\n```\n", false, "Code 1\nCode 2\n"},
		{"End",
			"One\n```\nCode 1\nCode 2\n```\nEnd\n", false, "Code 1\nCode 2\n"},
	}

	for i, c := range cs {
		st.proc(c.line)
		if st.markdown.String() != c.markdown {
			t.Errorf("Index %d: Expected markdown %q but got %q",
				i, c.markdown, st.markdown.String())
		}
		if st.inChunk != c.inChunk {
			t.Errorf("Index %d: Expected inChunk=%v but got %v",
				i, c.inChunk, st.inChunk)
		}
		if st.code.String() != c.code {
			t.Errorf("Index %d: Expected markdown %q but got %q",
				i, c.code, st.code.String())
		}
	}
}
