package main

import (
	"testing"
)

func TestSimpleOutName(t *testing.T) {
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
		act := simpleOutName(d.in)
		if act != d.exp {
			t.Errorf("Input %q, expected %q but got %q", d.in, d.exp, act)
		}
	}
}

func TestChapterOutName(t *testing.T) {
	data := []struct {
		outDir      string
		foundInName string
		exp         string
	}{
		// Out dir is current dir, book in current dir
		{".", "first.md", "first.html"},
		{".", "chaps/second.md", "chaps/second.html"},
		// Out dir is current dir, book is elsewhere
		{".", "first.md", "first.html"},
		{".", "chaps/second.md", "chaps/second.html"},
		// Some out dir, book in current dir
		{"odir", "first.md", "odir/first.html"},
		{"odir", "chaps/second.md", "odir/chaps/second.html"},
		// Out dir is current dir, book is elsewhere
		{"odir", "first.md", "odir/first.html"},
		{"odir", "chaps/second.md", "odir/chaps/second.html"},
	}

	for _, d := range data {
		actual := chapterOutName(d.outDir, d.foundInName)
		if actual != d.exp {
			t.Errorf("outName(%q, %q) got %q, expected %q",
				d.outDir, d.foundInName, actual, d.exp)
		}
	}
}
