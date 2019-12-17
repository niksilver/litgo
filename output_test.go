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
		{"aa/bb/c.md", "c.html"},
	}

	for _, d := range data {
		act := outName(d.in)
		if act != d.exp {
			t.Errorf("Input %q, expected %q but got %q", d.in, d.exp, act)
		}
	}
}
