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
	// More state fields
	inChunk bool
	code    strings.Builder
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

func (s *state) proc(line string) {
	// Collect lines in code chunks
	if s.inChunk && line == "```" {
		s.inChunk = false
	} else if s.inChunk {
		s.code.WriteString(line + "\n")
	} else if !s.inChunk && line == "```" {
		s.inChunk = true
	}

	// Do process other lines
	s.markdown.WriteString(line + "\n")

}
