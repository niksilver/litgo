// main.go
package main

import (
	// Imports
	"bufio"
	"fmt"
	"github.com/russross/blackfriday"
	"strings"
)

// Package level declarations
type processing interface {
	proc(string)
}

type state struct {
	markdown strings.Builder
	lineNum  int
	// More state fields
	inChunk   bool
	chunkName string
	code      map[string]strings.Builder
}

// Functions
func main() {
	input := []byte("# Hello world\n\nThis is my other literate document")
	output := blackfriday.Run(input)
	fmt.Println(string(output))
}

func processContent(c []byte, p processing) {
	r := strings.NewReader(string(c))
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		p.proc(sc.Text())
	}
}

func newState() state {
	return state{
		// Required state initialisers
		code: make(map[string]strings.Builder),
	}
}

func (s *state) proc(line string) {
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
		s.code[s.chunkName] = strings.Builder{}
		s.inChunk = true
	}

	// Send surviving lines to markdown
	s.markdown.WriteString(line + "\n")

}
