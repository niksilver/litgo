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
	s := state{}
	s.setFirstInName("book.md")
	s.book = "book.md"
	d := newDoc()

	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "First line 1\nFirst line 2",
		"second.md": "Second line 1\nSecond line 2",
	}

	mockFileReader := func(inName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[inName])}, nil
	}

	firstPassForAll(&s, &d, proc, mockFileReader)

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
	s := state{}
	s.setFirstInName("not-a-book.md")
	d := newDoc()

	data := map[string]string{
		"not-a-book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md":  "First line 1\nFirst line 2",   // Should not read this
		"second.md": "Second line 1\nSecond line 2", // Should not read this
	}

	mockFileReader := func(inName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[inName])}, nil
	}

	firstPassForAll(&s, &d, proc, mockFileReader)

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
	s := state{}
	s.setFirstInName("book.md")
	s.book = "book.md"
	d := newDoc()

	data := map[string]string{
		"book.md":   `* [First chapter](first.md)`,
		"first.md":  `[Second link](second.md)`,
		"second.md": "Second line 1\nSecond line 2", // Should not read this
	}

	mockFileReader := func(inName string) (io.ReadCloser, error) {
		return stringReadCloser{strings.NewReader(data[inName])}, nil
	}

	firstPassForAll(&s, &d, proc, mockFileReader)

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
	s := state{}
	s.setFirstInName("../aaa/book.md")
	s.book = "../aaa/book.md"
	d := newDoc()

	data := map[string]string{
		"../aaa/book.md": `* [First chapter](chaps/first.md)
             * [Second chapter](chaps/second.md)`,
		"../aaa/chaps/first.md":  "First line 1\nFirst line 2",
		"../aaa/chaps/second.md": "Second line 1\nSecond line 2",
	}

	mockFileReader := func(inName string) (io.ReadCloser, error) {
		content, okay := data[inName]
		if !okay {
			return nil, fmt.Errorf("No content found for key %q", inName)
		}
		return stringReadCloser{strings.NewReader(content)}, nil
	}

	err := firstPassForAll(&s, &d, proc, mockFileReader)
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

	first := d.markdown["../aaa/chaps/first.md"].String()
	if first != data["../aaa/chaps/first.md"]+"\n" {
		t.Errorf("Expected first.md markdown to be %q but got %q",
			data["../aaa/chaps/first.md"]+"\n", first)
	}

	second := d.markdown["../aaa/chaps/second.md"].String()
	if second != data["../aaa/chaps/second.md"]+"\n" {
		t.Errorf("Expected second.md markdown to be %q but got %q",
			data["../aaa/chaps/second.md"]+"\n", second)
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
