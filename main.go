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

type lattice struct {
	childrenOf map[string]set
	parentsOf  map[string]set
}

type cyclicError struct {
	chunks []string
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

func compileLattice(chunks map[string]string) lattice {
	tr := lattice{
		childrenOf: make(map[string]set),
		parentsOf:  make(map[string]set),
	}

	for name, content := range chunks {
		// Make sure this parent is in the lattice
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

			// Make sure this child is in the lattice
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

func topLevelChunksAreFilenames(lat lattice) error {
	badNames := make([]string, 0)
	for ch, pars := range lat.parentsOf {
		if len(pars) == 0 && !isFilename(ch) {
			badNames = append(badNames, ch)
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

func (e *cyclicError) Error() string {
	return "Found cyclic chunks: " +
		strings.Join(e.chunks, " -> ")
}

func errorIfCyclic(lat lattice) error {
	// Find the top level chunks
	top := make([]string, 0)
	for ch, pars := range lat.parentsOf {
		if len(pars) == 0 {
			top = append(top, ch)
		}
	}

	// Make a singleton list of these, which is our initial list of paths
	paths := make([][]string, 0)
	for _, par := range top {
		paths = append(paths, []string{par})
	}

	// As long as we've got some existing paths...
	for len(paths) > 0 {
		// New paths, initially none
		nPaths := make([][]string, 0)

		// For each existing path...
		for _, path := range paths {
			// Pick the last element and find its children
			lastElt := path[len(path)-1]
			chs := make([]string, 0)
			for key, _ := range lat.childrenOf[lastElt] {
				chs = append(chs, key)
			}

			// If there are no children, go on to the next path
			if len(chs) == 0 {
				continue
			}

			// Terminate with an error if the appears earlier in the path
			for i := 0; i < len(path)-1; i++ {
				if path[i] == lastElt {
					return &cyclicError{path[i:]}
				}
			}

			// Add our list of new paths. One new path for each child
			for _, ch := range chs {
				nPath := append(path, ch)
				nPaths = append(nPaths, nPath)
			}
		}

		// Our list of new paths becomes the list of paths to work on
		paths = nPaths
	}

	// If we've got here, then there are no cycles
	return nil
}
