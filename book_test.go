package main

import (
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
	s.setInName("book.md")
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
	s.setInName("not-a-book.md")
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
