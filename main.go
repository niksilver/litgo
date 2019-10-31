// main.go
package main

import (
	// Imports
	"bufio"
	"fmt"
	"github.com/gomarkdown/markdown"
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
