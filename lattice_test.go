package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestBasicLattice(t *testing.T) {
	chunks := map[string]string{
		"top":         "top one\n@{chunk two}\n@{chunk three}\ntop four",
		"chunk three": "three A\nthree B",
		"chunk two":   "Two A\n  @{chunk three}  \n@chunk three",
	}
	expected := lattice{
		childrenOf: map[string]set{
			"top":         {"chunk two": true, "chunk three": true},
			"chunk two":   {"chunk three": true},
			"chunk three": {},
		},
		parentsOf: map[string]set{
			"top":         {},
			"chunk two":   {"top": true},
			"chunk three": {"top": true, "chunk two": true},
		},
	}

	lat := compileLattice(chunks)
	if !reflect.DeepEqual(lat, expected) {
		t.Errorf("Basic lattices not equal. Expected\n%v\nGot\n%v",
			expected, lat)
	}
}

func TestReferredChunkName(t *testing.T) {
	names := [][]string{
		{"First line", ""},                  // Nothing
		{"Some @{Second line}", ""},         // Nothing because content before
		{"@{Third line} here", ""},          // Nothing because content after
		{"@{Fourth line}", "Fourth line"},   // Name
		{"  @{Fifth line}  ", "Fifth line"}, // Name because we ignore outer spaces
		{"@{  Sixth line  }", "Sixth line"}, // Name because we ignore inner spaces
	}

	for _, nm := range names {
		act := referredChunkName(nm[0])
		if act != nm[1] {
			t.Errorf("Line %q: Expected %q, but got %q",
				nm[0], nm[1], act)
		}
	}
}

func TestIsFilename(t *testing.T) {
	data := []struct {
		string
		bool
	}{
		{"", false},            // No dot, and empty
		{"aa", false},          // No dot, non-empty
		{"aa bb cc", false},    // No dot, even with several words
		{"a.", false},          // No text after dot
		{".a c", false},        // Space after the dot
		{".gitignore ", false}, // Space after the dot, at the end
		{".abc", true},
		{"note.txt", true},
		{".txt.bak", true},
	}

	for _, d := range data {
		act := isFilename(d.string)
		if act != d.bool {
			t.Errorf("Is %q a filename? Expected %v but got %v",
				d.string, d.bool, act)
		}
	}
}

func TestTopLevelChunksAreFilenames(t *testing.T) {
	// This lattice has one top level filename (good) but one not
	lat1 := lattice{
		parentsOf: map[string]set{
			"Top1":       {},
			"Middle":     {"Top1": true},
			"Middle.txt": {"Top1.txt": true},
			"Top.txt":    {},
		},
	}
	err1 := topLevelChunksAreFilenames(lat1)
	if err1 == nil || !strings.Contains(err1.Error(), "Top1") {
		t.Errorf("Lattice 1 should report error about 'Top1' but got error %q", err1)
	}

	// This lattice has two top level chunks, neither a filename
	lat2 := lattice{
		parentsOf: map[string]set{
			"Top2-1":     {},
			"Middle":     {"Top1": true},
			"Top2-2":     {},
			"Middle.txt": {"Top1.txt": true},
		},
	}
	err2 := topLevelChunksAreFilenames(lat2)
	if err2 == nil || !strings.Contains(err2.Error(), "Top2-1") {
		t.Errorf("Lattice 2 should report error about 'Top2-1' but got error %q", err2)
	}
	if err2 == nil || !strings.Contains(err2.Error(), "Top2-2") {
		t.Errorf("Lattice 2 should report error about 'Top2-2' but got error %q", err2)
	}

	// This lattice has two top level chunks, both filenames, so all is good
	lat3 := lattice{
		parentsOf: map[string]set{
			"Top3-1.txt": {},
			"Middle":     {"Top1": true},
			"Middle.txt": {"Top1.txt": true},
			"Top3-2.txt": {},
		},
	}
	err3 := topLevelChunksAreFilenames(lat3)
	if err3 != nil {
		t.Errorf("Lattice 3 should report no errors but got error %q", err2)
	}
}

func TestErrorIfCyclic1NotCyclic(t *testing.T) {
	// Check this lattice isn't cyclic:
	//
	// aa -- hh
	//    \
	// bb--- cc
	//   \
	//    - dd --------- ee -- ff
	//        \        /
	//         --- gg -

	lat := lattice{
		parentsOf: map[string]set{
			"aa": {},
			"bb": {},
			"cc": {"aa": true, "bb": true},
			"dd": {"bb": true},
			"ee": {"dd": true, "gg": true},
			"ff": {"ee": true},
			"gg": {"dd": true},
			"hh": {"aa": true},
		},
		childrenOf: map[string]set{
			"aa": {"hh": true, "cc": true},
			"bb": {"cc": true, "dd": true},
			"cc": {},
			"dd": {"ee": true, "gg": true},
			"ee": {"ff": true},
			"ff": {},
			"gg": {"ee": true},
			"hh": {},
		},
	}

	err := errorIfCyclic(lat)
	if err != nil {
		t.Errorf("Good lattice incorrectly found to be cyclic. Got error %q",
			err.Error())
	}
}

func TestErrorIfCyclic2EmptyLatticeNotCyclic(t *testing.T) {
	lat := lattice{
		parentsOf:  make(map[string]set, 0),
		childrenOf: make(map[string]set, 0),
	}

	err := errorIfCyclic(lat)
	if err != nil {
		t.Errorf("Empty lattice incorrectly found to be cyclic. Got error %q",
			err.Error())
	}
}

func TestErrorIfCyclic3IsCyclic(t *testing.T) {
	// Check this lattice is cyclic:
	//
	// aa -- hh
	//    \
	// bb--- cc
	//   \
	//    - dd --------- ee -- ff --> dd (a cycle)
	//        \        /
	//         --- gg -

	lat := lattice{
		parentsOf: map[string]set{
			"aa": {},
			"bb": {},
			"cc": {"aa": true, "bb": true},
			"dd": {"bb": true, "ff": true},
			"ee": {"dd": true, "gg": true},
			"ff": {"ee": true},
			"gg": {"dd": true},
			"hh": {"aa": true},
		},
		childrenOf: map[string]set{
			"aa": {"hh": true, "cc": true},
			"bb": {"cc": true, "dd": true},
			"cc": {},
			"dd": {"ee": true, "gg": true},
			"ee": {"ff": true},
			"ff": {"dd": true},
			"gg": {"ee": true},
			"hh": {},
		},
	}

	err := errorIfCyclic(lat)
	if err == nil {
		t.Errorf("Cyclic lattice not recognised")
	}

	subs := "dd -> ee -> ff -> dd"
	if !strings.Contains(err.Error(), subs) {
		t.Errorf("Cyclic lattice gave wrong error. Expected %q but got %q",
			subs, err.Error())
	}
}
