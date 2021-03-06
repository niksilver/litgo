package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestProcForMarkdown(t *testing.T) {
	s := newState()
	s.setFirstInName("markdown.md")
	d := newDoc()
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
		s.proc(&s, &d, c.line)
		mdb := d.markdown[s.inName]
		if mdb.String() != c.markdown {
			t.Errorf("Line %d: Expected markdown %q but got %q",
				i+1, c.markdown, mdb.String())
		}
	}
}

func TestProcForInChunks(t *testing.T) {
	s := newState()
	s.setFirstInName("in.md")
	d := newDoc()
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
		s.proc(&s, &d, c.line)
		if s.inChunk != c.inChunk {
			t.Errorf("Line %d: Expected inChunk=%v but got %v",
				i+1, c.inChunk, s.inChunk)
		}
	}
}

func TestProcForChunkNames(t *testing.T) {
	d := newDoc()
	s := newState()
	s.setFirstInName("names.md")
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
	exp := map[string][]string{
		"First":  []string{"Code line 1", "Code line 2"},
		"Second": []string{"Code line 3"},
	}

	for _, line := range lines {
		s.proc(&s, &d, line)
	}

	if len(d.chunks) != 2 {
		t.Errorf("Expected 2 chunks but got %d: %#v",
			len(d.chunks), d.chunks)
	}
	for name, codes := range exp {
		if _, okay := d.chunks[name]; !okay {
			t.Errorf("Couldn't find chunk name %s", name)
			continue
		}
		cont := d.chunks[name].cont
		if len(cont) != len(codes) {
			t.Errorf("In chunk %s expected %d lines of content but content is %#v",
				name, len(codes), cont)
		}
		for i, code := range codes {
			if cont[i].code != code {
				t.Errorf("In chunk %s expected code[%d] == %q but got %q",
					name, i, code, cont[i].code)
			}
		}
	}
}

func TestProcForChunkDetails(t *testing.T) {
	s := newState()
	s.setFirstInName("details.md")
	d := newDoc()
	lines := []string{
		"``` First",
		"Code line 1",
		"Code line 2",
		"```",
		"# Heading",
		"``` Second",
		"Code line 3",
		"```",
		"",
		"``` First", // Appending to a chunk
		"Code line 4",
		"```",
		"The end",
	}
	sec0 := section{s.inName, []int(nil), ""}
	sec1 := section{s.inName, []int{1}, "Heading"}
	expected := map[string]chunk{
		"First": chunk{
			[]chunkDef{
				chunkDef{"details.md", 1, sec0},
				chunkDef{"details.md", 10, sec1},
			},
			[]chunkCont{
				chunkCont{"details.md", 2, "Code line 1"},
				chunkCont{"details.md", 3, "Code line 2"},
				chunkCont{"details.md", 11, "Code line 4"},
			},
		},
		"Second": chunk{
			[]chunkDef{
				chunkDef{"details.md", 6, sec1},
			},
			[]chunkCont{
				chunkCont{"details.md", 7, "Code line 3"},
			},
		},
	}

	for _, line := range lines {
		s.proc(&s, &d, line)
	}

	if len(d.chunks) != 2 {
		t.Errorf("Expected 2 chunks but got %d", len(d.chunks))
	}
	if !reflect.DeepEqual(expected["First"], *d.chunks["First"]) {
		t.Errorf("Expected First chunk to be\n%#v\nbut got\n%#v",
			expected["First"], *d.chunks["First"])
	}
	if !reflect.DeepEqual(expected["Second"], *d.chunks["Second"]) {
		t.Errorf("Expected Second chunk to be\n%#v\nbut got\n%#v",
			expected["Second"], *d.chunks["Second"])
	}
}

func TestProcForWarningsAroundChunks(t *testing.T) {
	s := newState()
	s.setFirstInName("testfile.lit")
	d := newDoc()
	lines := []string{
		"Title",
		"",
		"``` Okay chunk",
		"Chunk content",
		"```",
		"",
		"```", // Chunk start without name
		"```",
		"",
		"``` Another chunk",
		"Chunk content", // Chunk does not end
	}
	r := strings.NewReader(strings.Join(lines, "\n"))
	expected := []struct {
		fName string
		line  int
		subs  string
	}{
		{"testfile.lit", 7, "no name"},
		{"testfile.lit", 11, "chunk not closed"},
	}

	processContent(r, &s, &d)

	nWarn := len(s.warnings)
	if nWarn != len(expected) {
		t.Errorf("Expected %d warnings, but got %d", len(expected), nWarn)
	}
	for i, w := range expected {
		if i+1 > nWarn {
			t.Errorf("Warning index %d missing, expected %v", i, w)
			continue
		}
		if expected[i].line != s.warnings[i].line ||
			expected[i].fName != s.warnings[i].fName ||
			!strings.Contains(s.warnings[i].msg, expected[i].subs) {
			t.Errorf("Expected warning index %d to be %v but got %v",
				i, w, s.warnings)
		}
	}
}

func TestProcForChunkRefs(t *testing.T) {
	s := newState()
	s.setFirstInName("testfile.lit")
	d := newDoc()
	lines := []string{
		"Opening text", // Line 1
		"",
		"``` Chunk one",
		"Chunk content",
		"```", // Line 5
		"# First section",
		"``` Chunk two",
		"# Comment, not section heading",
		"```", // Line 9
		"",
		"``` Chunk three",
		"More chunk content",
		"```", // Line 13
	}
	sec0 := section{s.inName, []int(nil), ""}
	sec1 := section{s.inName, []int{1}, "First section"}
	r := strings.NewReader(strings.Join(lines, "\n"))
	expected := map[int]chunkRef{
		5:  chunkRef{"Chunk one", sec0},
		9:  chunkRef{"Chunk two", sec1},
		13: chunkRef{"Chunk three", sec1},
	}

	processContent(r, &s, &d)

	chRefs := d.chunkRefs[s.inName]
	if len(chRefs) != len(expected) {
		t.Errorf("Expected %d chunk refs but got %d. Map is %#v",
			len(expected), len(chRefs), chRefs)
		return
	}
	for lNum, ref := range expected {
		if !reflect.DeepEqual(chRefs[lNum], ref) {
			t.Errorf("For line %d expected chunk %#v but got %#v",
				lNum, ref, chRefs[lNum])
		}
	}
}
