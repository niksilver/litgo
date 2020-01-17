package main

import (
	"io"
	"strings"
	"testing"
)

func TestFinalMarkdown_AmendChapterLinks(t *testing.T) {
	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](sec/second.md)`,
		"first.md": `* First [line 1](book.md)
            * First [line 2](sec/second.md)
            * First line 3`,
		"sec/second.md": `* Second line 1\
            * Second [line 2](not-a-chapter.md)
            * [First chapter](../first.md)`,
	}
	expected := map[string][]string{
		"book.md": []string{
			"[First chapter](first.html)",
			"[Second chapter](sec/second.html)",
		},
		"first.md": []string{
			"[line 1](book.html)",
			"[line 2](sec/second.html)",
		},
		"sec/second.md": []string{
			"[line 2](not-a-chapter.md)",
			"[First chapter](../first.html)",
		},
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)

	for inName, subs := range expected {
		if _, okay := d.markdown[inName]; !okay {
			t.Errorf("No markdown for inName %s", inName)
			continue
		}
		mdown := finalMarkdown(inName, &d).String()
		for _, sub := range subs {
			if !strings.Contains(mdown, sub) {
				t.Errorf("Input %s did not contain %q. Content is\n%s",
					inName, sub, mdown)
			}
		}
	}
}

func TestFinalMarkdown_AmendedChapterLinksCorrectWhenBookInDir(t *testing.T) {
	data := map[string]string{
		"../aaa/book.md": `* [First chapter](chap/first.md)
             * [Second chapter](chap/second.md)`,
		"../aaa/chap/first.md": `First line 1
            First [line 2](second.md)
            First line 3`,
		"../aaa/chap/second.md": `Second [line 1](../book.md)
            Second [line 2](not-a-chapter.md)
            Second line 3`,
	}
	expected := map[string][]string{
		"../aaa/book.md": []string{
			"[First chapter](chap/first.html)",
			"[Second chapter](chap/second.html)",
		},
		"../aaa/chap/first.md": []string{
			"[line 2](second.html)",
		},
		"../aaa/chap/second.md": []string{
			"[line 1](../book.html)",
			"[line 2](not-a-chapter.md)",
		},
	}

	s := newState()
	s.setFirstInName("../aaa/book.md")
	s.book = "../aaa/book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)

	for inName, subs := range expected {
		if _, okay := d.markdown[inName]; !okay {
			t.Errorf("No markdown for inName %s", inName)
			continue
		}
		mdown := finalMarkdown(inName, &d).String()
		for _, sub := range subs {
			if !strings.Contains(mdown, sub) {
				t.Errorf("Input %s did not contain %q. Content is\n%s",
					inName, sub, mdown)
			}
		}
	}
}

func TestFinalMarkdown_AmendChapterLinks_NotInChunks(t *testing.T) {
	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](sec/second.md)`,
		"first.md": "``` Chunk one\n" +
			"First [line 2](sec/second.md)\n" +
			"```\n",
		"sec/second.md": `* Second line 1\
            * Second [line 2](not-a-chapter.md)
            * [First chapter](../first.md)`,
	}
	expected := map[string][]string{
		"book.md": []string{
			"[First chapter](first.html)",
			"[Second chapter](sec/second.html)",
		},
		"first.md": []string{
			"[line 2](sec/second.md)", // Not changed as it's in a chunk
		},
		"sec/second.md": []string{
			"[line 2](not-a-chapter.md)",
			"[First chapter](../first.html)",
		},
	}

	s := newState()
	s.setFirstInName("book.md")
	s.book = "book.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)

	for inName, subs := range expected {
		if _, okay := d.markdown[inName]; !okay {
			t.Errorf("No markdown for inName %s", inName)
			continue
		}
		mdown := finalMarkdown(inName, &d).String()
		for _, sub := range subs {
			if !strings.Contains(mdown, sub) {
				t.Errorf("Input %s did not contain %q. Content is\n%s",
					inName, sub, mdown)
			}
		}
	}
}
