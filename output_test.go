package main

import (
	"testing"
)

func TestOutName(t *testing.T) {
	data := []struct {
		in  string
		exp string
	}{
		{"docs.md", "docs.html"},
		{"doc", "doc.html"},
		{"-", "out.html"},
		{"", "out.html"},
		{"aa/bb/c.md", "aa/bb/c.html"},
		{"../bb/c.md", "../bb/c.html"},
	}

	for _, d := range data {
		act := outName(d.in)
		if act != d.exp {
			t.Errorf("Input %q, expected %q but got %q", d.in, d.exp, act)
		}
	}
}

func TestOutNames(t *testing.T) {
	data := []struct {
		outDir  string
		inNames []string
		exp     []string
	}{
		{"",
			[]string{"../aaa/book.md", "sub/ch1.md", "sub/ch2.md"},
			[]string{"../aaa/book.html", "../aaa/sub/ch1.html", "../aaa/sub/ch2.html"},
		},
		{"docs",
			[]string{"../aaa/book.md", "sub/ch1.md", "sub/ch2.md"},
			[]string{"docs/book.html", "docs/sub/ch1.html", "docs/sub/ch2.html"},
		},
	}

	for _, d := range data {
		actNames := outNames(d.outDir, d.inNames)
		for i, actName := range actNames {
			if actName != d.exp[i] {
				t.Errorf("our dir = %q, book in = %q, for %q got %q but expected %q",
					d.outDir, d.inNames[0], d.inNames[i], actName, d.exp[i])
			}
		}
	}
}
