package main

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type stringReadCloser struct {
	io.Reader
}

func (src stringReadCloser) Close() error {
	return nil
}

func TestReadBookAndChapters_FollowsLinks(t *testing.T) {
	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "First line 1\nFirst line 2",
		"second.md": "Second line 1\nSecond line 2",
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(inName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[inName])}, nil
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

func TestReadBookAndChapters_DontFollowsLinksIfNotBook(t *testing.T) {
	data := map[string]string{
		"not-a-book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "First line 1\nFirst line 2",   // Should not read this
		"second.md": "Second line 1\nSecond line 2", // Should not read this
	}

	s := newState()
	s.setFirstInName("not-a-book.md")
	s.reader = func(inName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[inName])}, nil
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

func TestReadBookAndChapters_DontFollowsLinksBelowBookLevel(t *testing.T) {
	data := map[string]string{
		"book.md":   `* [First chapter](first.md)`,
		"first.md":  `[Second link](second.md)`,
		"second.md": "Second line 1\nSecond line 2", // Should not read this
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(inName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[inName])}, nil
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

func TestReadBookAndChapters_FollowsLinksWhenBookNotInBaseDir(t *testing.T) {
	data := map[string]string{
		"../aaa/book.md": `* [First chapter](chaps/first.md)
             * [Second chapter](chaps/second.md)`,
		"chaps/first.md":  "First line 1\nFirst line 2",
		"chaps/second.md": "Second line 1\nSecond line 2",
	}

	s := newState()
	s.setFirstInName("../aaa/book.md")
	s.book = "../aaa/book.md"
	s.reader = func(inName string) (io.ReadCloser, error) {
		content, okay := data[inName]
		if !okay {
			return nil, fmt.Errorf("No content found for key %q", inName)
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

	first := d.markdown["chaps/first.md"].String()
	if first != data["chaps/first.md"]+"\n" {
		t.Errorf("Expected first.md markdown to be %q but got %q",
			data["chaps/first.md"]+"\n", first)
	}

	second := d.markdown["chaps/second.md"].String()
	if second != data["chaps/second.md"]+"\n" {
		t.Errorf("Expected second.md markdown to be %q but got %q",
			data["chaps/second.md"]+"\n", second)
	}
}

func TestReadBookAndChapters_WriteToMarkdownOutDir(t *testing.T) {
	data := map[string]string{
		"../aaa/book.md": `* [First chapter](chaps/first.md)
             * [Second chapter](chaps/second.md)`,
		"chaps/first.md":  "First line 1\nFirst line 2",
		"chaps/second.md": "Second line 1\nSecond line 2",
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
	s.reader = func(inName string) (io.ReadCloser, error) {
		content, okay := data[inName]
		if !okay {
			return nil, fmt.Errorf("No content found for key %q", inName)
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
			t.Errorf("Did not expect markdown for file %s", name)
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
