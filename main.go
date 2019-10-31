// main.go
package main

import (
	// Imports
	"bufio"
	"fmt"
	"github.com/gomarkdown/markdown"
	"regexp"
	"strings"
)

// Package level declarations
type state struct {
	markdown strings.Builder
	lineNum  int
	// More state fields
	warnings  []problem
	inChunk   bool
	chunkName string
	code      map[string]strings.Builder

	proc func(s *state, line string)
}

type problem struct {
	line int
	msg  string
}

type set map[string]bool

type tree struct {
	childrenOf map[string]set
	parentsOf  map[string]set
}

// Functions
func newState() state {
	return state{
		// Field initialisers for state
		proc: proc,
		code: make(map[string]strings.Builder),
	}
}

func processContent(c []byte, s *state) {
	r := strings.NewReader(string(c))
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		s.proc(s, sc.Text())
	}
	// Tidy-up after processing content
	if s.inChunk {
		s.warnings = append(s.warnings, problem{s.lineNum, "Content finished but chunk not closed"})
	}

}

func main() {
	input := []byte("# Hello world\n\nThis is my other literate document")
	s := newState()
	processContent(input, &s)
	md := []byte(s.markdown.String())
	output := markdown.ToHTML(md, nil, nil)
	fmt.Println(string(output))
	//@{Check code chunks}
}

func proc(s *state, line string) {
	s.lineNum++
	// Collect lines in code chunks
	if s.inChunk && line == "```" {
		s.inChunk = false
	} else if s.inChunk {
		b := s.code[s.chunkName]
		b.WriteString(line + "\n")
		s.code[s.chunkName] = b
	} else if !s.inChunk && strings.HasPrefix(line, "```") {
		s.chunkName = strings.TrimSpace(line[3:])
		if s.chunkName == "" {
			s.warnings = append(s.warnings, problem{s.lineNum, "Chunk has no name"})
		}
		s.code[s.chunkName] = strings.Builder{}
		s.inChunk = true
	}

	// Send surviving lines to markdown
	s.markdown.WriteString(line + "\n")

}

func compileTree(chunks map[string]string) tree {
	tr := tree{
		childrenOf: make(map[string]set),
		parentsOf:  make(map[string]set),
	}

	for name, content := range chunks {
		// Make sure this parent is in the tree
		if tr.childrenOf[name] == nil {
			tr.childrenOf[name] = make(map[string]bool)
		}
		if tr.parentsOf[name] == nil {
			tr.parentsOf[name] = make(map[string]bool)
		}

		sc := bufio.NewScanner(strings.NewReader(content))
		for sc.Scan() {
			line := sc.Text()
			refChunk := referredChunkName(line)
			if refChunk == "" {
				continue
			}

			// Make sure this child is in the tree
			if tr.childrenOf[refChunk] == nil {
				tr.childrenOf[refChunk] = make(map[string]bool)
			}
			if tr.parentsOf[refChunk] == nil {
				tr.parentsOf[refChunk] = make(map[string]bool)
			}

			// Store the parent/child relationship
			(tr.childrenOf[name])[refChunk] = true
			(tr.parentsOf[refChunk])[name] = true
		}
	}
	return tr
}

func referredChunkName(str string) string {
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "@{") && strings.HasSuffix(str, "}") {
		return strings.TrimSpace(str[2 : len(str)-1])
	}
	return ""
}

func topLevelChunksAreFilenames(tr tree) error {
	badNames := make([]string, 0)
	for par, chs := range tr.parentsOf {
		if len(chs) == 0 {
			if !isFilename(par) {
				badNames = append(badNames, par)
			}
		}
	}

	if len(badNames) == 0 {
		// No error
		return nil
	}

	msg := "Found top level chunk which isn't a filename: %s"
	if len(badNames) > 1 {
		msg = "Found top level chunks which aren't filenames: %s"
	}
	return fmt.Errorf(msg, strings.Join(badNames, ","))
}

func isFilename(s string) bool {
	match, _ := regexp.MatchString("\\.\\S+$", s)
	return match
}
