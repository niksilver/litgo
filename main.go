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
	// Ignore lines that start with X
	if strings.HasPrefix(line, "X") {
		return
	}

	// Do process other lines
	s.markdown.WriteString(line)

}
