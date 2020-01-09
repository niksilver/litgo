package main

import (
	"io"
	"regexp"
	"strings"
	"testing"
)

func TestFinalMarkdown_AmendChapterLinks(t *testing.T) {
	data := map[string]string{
		"book.md": `* [First chapter](first.md)
             * [Second chapter](second.md)`,
		"first.md": `First line 1
            First [line 2](second.md)
            First line 3`,
		"second.md": `Second line 1\
            Second [line 2](not-a-chapter.md)
            Second line 3`,
	}
	expected := map[string][]string{
		"book.md": []string{
			"\\[First chapter\\]\\(first.html\\)",
			"\\[Second chapter\\]\\(second.html\\)",
		},
		"first.md": []string{
			"\\[line 2\\]\\(second.html\\)",
		},
		"second.md": []string{
			"\\[line 2\\]\\(not-a-chapter.md\\)",
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

	for inName, reStrs := range expected {
		if _, okay := d.markdown[inName]; !okay {
			t.Errorf("No markdown for inName %s", inName)
			continue
		}
		mdown := finalMarkdown(inName, &d).String()
		for _, reStr := range reStrs {
			match, err := regexp.MatchString(reStr+"(?m)", mdown)
			if err != nil {
				t.Errorf("Problem with regexp %q: %s", reStr, err.Error())
				continue
			}
			if !match {
				t.Errorf("Input %s did not match regexp %q. Content is\n%s",
					inName, reStr, mdown)
			}
		}
	}
}
