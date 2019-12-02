// main.go
package main

import (
	// Imports
	"bufio"
	"fmt"
	"github.com/gomarkdown/markdown"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Package level declarations
type state struct {
	fname    string          // Name of file being processed, relative to working dir
	markdown strings.Builder // Markdown output
	lineNum  int
	// More state fields
	warnings  []warning
	inChunk   bool              // If we're currently reading a chunk
	chunkName string            // Name of current chunk
	chunks    map[string]*chunk // All the chunks found so far
	lineDir   string
}

type warning struct {
	fname string
	line  int
	msg   string
}

type chunk struct {
	line  []int    // Line numbers where the chunk is defined
	code  []string // Lines of code, without newlines
	lines []int    // Line number for each line of code
}

type set map[string]bool

type lattice struct {
	childrenOf map[string]set
	parentsOf  map[string]set
}

type section struct {
	nums []int
	text string
}

// Functions
func newState() state {
	return state{
		// Field initialisers for state
		chunks: make(map[string]*chunk),
	}
}

func processContent(c []byte, s *state, proc func(*state, string)) {
	r := strings.NewReader(string(c))
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		proc(s, sc.Text())
	}
	// Tidy-up after processing content
	if s.inChunk {
		s.warnings = append(s.warnings,
			warning{s.fname, s.lineNum, "Content finished but chunk not closed"})
	}

}

func main() {
	s := newState()
	input, err := inputBytes(s.fname)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	processContent(input, &s, proc)

	// Check code chunks and maybe abort
	lat := compileLattice(s.chunks)
	errs := make([]error, 0)
	if err := assertTopLevelChunksAreFilenames(lat); err != nil {
		errs = append(errs, err)
	}
	if err := assertNoCycles(lat); err != nil {
		errs = append(errs, err)
	}
	if err := assertAllChunksDefined(s.chunks, lat); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println(e.Error())
		}
		return
	}

	// Write out warnings
	for _, w := range s.warnings {
		fmt.Printf("%s: %d: %s\n", w.fname, w.line, w.msg)
	}

	// Write out code chunks
	top := topLevelChunks(lat)
	err = writeChunks(top, s, makeChunkWriter)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	md := []byte(s.markdown.String())
	output := markdown.ToHTML(md, nil, nil)
	fmt.Println(string(output))
}

func inputBytes(fname string) (input []byte, e error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			e = err
		}
	}()
	return ioutil.ReadAll(f)
}

func proc(s *state, line string) {
	s.lineNum++
	// Collect lines in code chunks
	if s.inChunk && line == "```" {
		s.inChunk = false
	} else if s.inChunk {
		ch := s.chunks[s.chunkName]
		s.chunks[s.chunkName].code = append(ch.code, line)
		s.chunks[s.chunkName].lines = append(ch.lines, s.lineNum)
	} else if !s.inChunk && strings.HasPrefix(line, "```") {
		s.chunkName = strings.TrimSpace(line[3:])
		if s.chunkName == "" {
			s.warnings = append(s.warnings,
				warning{s.fname, s.lineNum, "Chunk has no name"})
		}
		ch := s.chunks[s.chunkName]
		if ch == nil {
			s.chunks[s.chunkName] = &chunk{}
			ch = s.chunks[s.chunkName]
		}
		s.chunks[s.chunkName].line = append(ch.line, s.lineNum)
		s.inChunk = true
	}

	// Send surviving lines to markdown
	s.markdown.WriteString(line + "\n")

}

func compileLattice(chunks map[string]*chunk) lattice {
	lat := lattice{
		childrenOf: make(map[string]set),
		parentsOf:  make(map[string]set),
	}

	for name, data := range chunks {
		// Make sure this parent is in the lattice
		if lat.childrenOf[name] == nil {
			lat.childrenOf[name] = make(map[string]bool)
		}
		if lat.parentsOf[name] == nil {
			lat.parentsOf[name] = make(map[string]bool)
		}

		for _, line := range data.code {
			refChunk := referredChunkName(line)
			if refChunk == "" {
				continue
			}

			// Make sure this child is in the lattice
			if lat.childrenOf[refChunk] == nil {
				lat.childrenOf[refChunk] = make(map[string]bool)
			}
			if lat.parentsOf[refChunk] == nil {
				lat.parentsOf[refChunk] = make(map[string]bool)
			}

			// Store the parent/child relationship
			(lat.childrenOf[name])[refChunk] = true
			(lat.parentsOf[refChunk])[name] = true
		}
	}
	return lat
}

func referredChunkName(str string) string {
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "@{") && strings.HasSuffix(str, "}") {
		return strings.TrimSpace(str[2 : len(str)-1])
	}
	return ""
}

func assertTopLevelChunksAreFilenames(lat lattice) error {
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

func assertNoCycles(lat lattice) error {
	// Find the top level chunks
	top := topLevelChunks(lat)

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

			// Terminate with an error if the elt appears earlier in the path
			for i := 0; i < len(path)-1; i++ {
				if path[i] == lastElt {
					return fmt.Errorf("Found cyclic chunks: %s",
						strings.Join(path[i:], " -> "))
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

func topLevelChunks(lat lattice) []string {
	top := make([]string, 0)
	for ch, pars := range lat.parentsOf {
		if len(pars) == 0 {
			top = append(top, ch)
		}
	}
	return top
}

func assertAllChunksDefined(chunks map[string]*chunk, lat lattice) error {
	missing := make([]string, 0)
	for par, _ := range lat.childrenOf {
		if chunks[par] == nil {
			missing = append(missing, par)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	s := ""
	if len(missing) >= 2 {
		s = "s"
	}
	return fmt.Errorf("Chunk%s not defined: %s",
		s, strings.Join(missing, ", "))
}

func writeChunks(top []string, s state, wf func(string) (io.StringWriter, error)) error {
	for _, name := range top {
		w, err := wf(name)
		if err != nil {
			return err
		}
		err = writeChunk(name, s, w, "")
		if err != nil {
			return err
		}
	}

	// No errors - all okay
	return nil
}

func makeChunkWriter(name string) (io.StringWriter, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return bufio.NewWriter(f), nil
}

func writeChunk(name string,
	s state,
	w io.StringWriter,
	indent string) error {

	chunk := *s.chunks[name]
	for i, code := range chunk.code {
		var err error
		if ref := referredChunkName(code); ref != "" {
			iPos := strings.Index(code, "@")
			err = writeChunk(ref, s, w, code[0:iPos]+indent)
		} else {
			lnum := chunk.lines[i]
			dir := lineDirective(s.lineDir, s.fname, lnum)
			_, err = w.WriteString(indent + dir + code + "\n")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func lineDirective(dir string, fname string, n int) string {
	if dir == "" {
		return ""
	}

	out := ""
	perc := false
	for _, r := range dir {
		if perc {
			switch r {
			case '%':
				out += "%"
			case 'f':
				out += fname
			case 'l':
				out += fmt.Sprintf("%d", n)
			default:
				out += string(r)
			}
			perc = false
		} else if r == '%' {
			perc = true
		} else {
			out += string(r)
			perc = false
		}
	}
	return out + "\n"
}

func (s *section) toString() string {
	if len(s.nums) == 0 {
		return ""
	}

	num := ""
	for _, n := range s.nums {
		num += strconv.Itoa(n) + "."
	}
	return num + " " + s.text
}

func (s *section) next(line string) section {
	return section{}
}
