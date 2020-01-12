package main

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

type stringReadCloser struct {
	io.Reader
}

func (src stringReadCloser) Close() error {
	return nil
}

func TestFirstPassForAll_FollowsLinks(t *testing.T) {
	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "First line 1\nFirst line 2",
		"second.md": "Second line 1\nSecond line 2",
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)

	if len(d.markdown) != 3 {
		t.Errorf("Expected 3 markdown docs but got %d: %#v",
			len(d.markdown), d.markdown)
		return
	}

	book := d.markdown["book.md"].String()
	if len(book) < 20 {
		t.Errorf("Markdown for book.md is too short. Got %q", book)
	}

	first := d.markdown["first.md"].String()
	if first != data["first.md"]+"\n" {
		t.Errorf("Expected first.md markdown to be %q but got %q",
			data["first.md"]+"\n", first)
	}

	second := d.markdown["second.md"].String()
	if second != data["second.md"]+"\n" {
		t.Errorf("Expected second.md markdown to be %q but got %q",
			data["second.md"]+"\n", second)
	}
}

func TestFirstPassForAll_DontFollowsLinksIfNotBook(t *testing.T) {
	data := map[string]string{
		"not-a-book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "First line 1\nFirst line 2",   // Should not read this
		"second.md": "Second line 1\nSecond line 2", // Should not read this
	}

	s := newState()
	s.setFirstInName("not-a-book.md")
	s.reader = func(fName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)

	if len(d.markdown) != 1 {
		t.Errorf("Expected 1 markdown doc but got %d: %#v",
			len(d.markdown), d.markdown)
		return
	}

	book := d.markdown["not-a-book.md"].String()
	if len(book) < 20 {
		t.Errorf("Markdown for not-a-book.md is too short. Got %q", book)
	}
}

func TestFirstPassForAll_DontFollowsLinksBelowBookLevel(t *testing.T) {
	data := map[string]string{
		"book.md":   `* [First chapter](first.md)`,
		"first.md":  `[Second link](second.md)`,
		"second.md": "Second line 1\nSecond line 2", // Should not read this
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)

	if len(d.markdown) != 2 {
		t.Errorf("Expected 2 markdown doc but got %d: %#v",
			len(d.markdown), d.markdown)
		return
	}

	book := d.markdown["book.md"].String()
	if len(book) < 20 {
		t.Errorf("Markdown for book.md is too short. Got %q", book)
	}

	first := d.markdown["first.md"].String()
	if first != data["first.md"]+"\n" {
		t.Errorf("Expected first.md markdown to be %q but got %q",
			data["first.md"]+"\n", first)
	}
}

func TestFirstPassForAll_FollowsLinksWhenBookNotInBaseDir(t *testing.T) {
	data := map[string]string{
		"../aaa/book.md": `* [First chapter](chaps/first.md)
             * [Second chapter](chaps/second.md)`,
		"../aaa/chaps/first.md":  "First line 1\nFirst line 2",
		"../aaa/chaps/second.md": "Second line 1\nSecond line 2",
	}

	s := newState()
	s.setFirstInName("../aaa/book.md")
	s.book = "../aaa/book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		content, okay := data[fName]
		if !okay {
			return nil, fmt.Errorf("No content found for file name %q", fName)
		}
		return stringReadCloser{strings.NewReader(content)}, nil
	}
	d := newDoc()

	err := firstPassForAll(&s, &d)
	if err != nil {
		t.Errorf("Error on first pass for all: %s", err.Error())
	}

	if len(d.markdown) != 3 {
		t.Errorf("Expected 3 markdown docs but got %d: %#v",
			len(d.markdown), d.markdown)
		return
	}

	book := d.markdown["../aaa/book.md"].String()
	if len(book) < 20 {
		t.Errorf("Markdown for ../aaa/book.md is too short. Got %q", book)
	}

	firstInName := "../aaa/chaps/first.md"
	first := d.markdown[firstInName].String()
	if first != data[firstInName]+"\n" {
		t.Errorf("Expected first.md markdown to be %q but got %q",
			data[firstInName]+"\n", first)
	}

	secondInName := "../aaa/chaps/second.md"
	second := d.markdown[secondInName].String()
	if second != data[secondInName]+"\n" {
		t.Errorf("Expected second.md markdown to be %q but got %q",
			data[secondInName]+"\n", second)
	}
}

func TestFirstPassForAll_PreservesSectionForNewChapter(t *testing.T) {
	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "# Section 1\n# Section 2\n## Section 2.1",
		"second.md": "Second line 1",
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	var sec section
	secSet := false
	secExp := section{"second.md", []int{2, 1}, "Section 2.1"}
	procOrig := s.proc
	s.proc = func(s *state, d *doc, line string) {
		if !secSet && s.inName == "second.md" {
			sec = s.sec
			secSet = true
		}
		procOrig(s, d, line)
	}
	d := newDoc()

	firstPassForAll(&s, &d)

	if !secSet {
		t.Errorf("Finished reading markup but sec was not set")
	}
	if !reflect.DeepEqual(sec, secExp) {
		t.Errorf("At start of new file, got section %#v but expected %#v",
			sec, secExp)
	}
}

func TestFirstPassForAll_WriteToMarkdownOutDir(t *testing.T) {
	data := map[string]string{
		"../aaa/book.md": `* [First chapter](chaps/first.md)
             * [Second chapter](chaps/second.md)`,
		"../aaa/chaps/first.md":  "First line 1\nFirst line 2",
		"../aaa/chaps/second.md": "Second line 1\nSecond line 2",
	}

	// Substrings we expect to see in the HTML
	expected := map[string]string{
		"outdir/book.html":         "First chapter",
		"outdir/chaps/first.html":  "First line 1",
		"outdir/chaps/second.html": "Second line 1",
	}

	s := newState()
	s.setFirstInName("../aaa/book.md")
	s.book = "../aaa/book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		content, okay := data[fName]
		if !okay {
			return nil, fmt.Errorf("No content found for file %q", fName)
		}
		return stringReadCloser{strings.NewReader(content)}, nil
	}

	d := newBuilderDoc(newDoc())
	d.docOutDir = "outdir"

	err1 := firstPassForAll(&s, &d.doc)
	if err1 != nil {
		t.Errorf("Error on first pass for all: %s", err1.Error())
	}

	err2 := writeAllMarkdown(s.inNames, &d.doc)
	if err2 != nil {
		t.Errorf("Error on writeAllMarkdown: %s", err2.Error())
	}

	if len(d.outputs) != 3 {
		t.Errorf("Expected 3 markdown docs but got %d: %#v",
			len(d.outputs), d.outputs)
		return
	}

	for name, sb := range d.outputs {
		if expected[name] == "" {
			t.Errorf("Did not expect markdown for file %q", name)
			continue
		}
		expSub := expected[name]
		act := sb.String()
		if !strings.Contains(act, expSub) {
			t.Errorf("File %s, expected to find substring %q but got string %q",
				name, expSub, act)
		}
	}
}

func TestFirstPassForAll_ErrorsIfInChunkAtEndOfFile(t *testing.T) {
	data := map[string]string{
		"book.md":  "* [First chapter](first.md)",
		"first.md": "First line 1\n``` Chunk one\n\n",
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	err := firstPassForAll(&s, &d)
	if err == nil {
		t.Errorf("Should have errored due to end of file within chunk")
		return
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "in chunk") {
		t.Errorf("Error should mention it's in a chunk, but error was: %s", errStr)
	}
	if !strings.Contains(errStr, "first.md") {
		t.Errorf("Error should mention input file first.md but was: %s", errStr)
	}
}

func TestMarkdownLink(t *testing.T) {
	data := []struct {
		line string
		link string
	}{
		{"", ""},
		{"no.md", ""},
		{"...](some/file.md)...", "some/file.md"},
		{"...](some/file.md...", ""},
		{"...](some/file.md \"Title\")", "some/file.md"},
		{"...](some/file.txt)...", ""},
	}

	for _, d := range data {
		actual := markdownLink(d.line)
		if actual != d.link {
			t.Errorf("For line %q expected link %q but got %q",
				d.line, d.link, actual)
		}
	}
}
