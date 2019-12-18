// main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gomarkdown/markdown"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Package level declarations
type state struct {
	// Tracking
	inName    string    // Name of file being processed, relative to working dir
	outName   string    // Name of final file to write to
	lineNum   int       // Current line number
	chunkName string    // Name of current chunk
	inChunk   bool      // If we're currently reading a chunk
	warnings  []warning // Warnings we're collecting
	sec       section   // Current section being read
}

type doc struct {
	markdown    map[string]*strings.Builder // Markdown after the initial read, per in file
	chunks      map[string]*chunk           // All the chunks found so far
	chunkStarts map[int]string              // Lines where a named chunk starts
	chunkRefs   map[int]chunkRef            // Lines where other chunks are called in
	lat         lattice                     // A lattice of chunk parent/child relationships
	secStarts   map[int]section             // Lines where a section starts
	// Config
	lineDir string // The string pattern for line directives
}

type lineProc = func(*state, *doc, string)
type warning struct {
	fName string
	line  int
	msg   string
}

type section struct {
	nums []int
	text string
}

type chunk struct {
	def  []chunkDef  // Each place where the chunk is defined
	cont []chunkCont // Each line of code
}

// Where the chunk is defined: file name, line number, section
type chunkDef struct {
	file string
	line int
	sec  section
}

// A line of chunk content: file name, line number, and the code line itself
type chunkCont struct {
	file string
	lNum int
	code string
}
type chunkRef struct {
	name    string
	thisSec section
}

type set map[string]bool

type lattice struct {
	childrenOf map[string]set
	parentsOf  map[string]set
}

var lDir string

// Functions

func init() {
	// Flag initialisation
	flag.StringVar(&lDir, "line-dir", "", "Pattern for line directives")

}

func main() {
	// Set up the initial state
	s := state{}
	d := newDoc()

	// Update the structs according to the command line
	flag.Parse()
	if flag.NArg() == 0 {
		s.inName = "-"
	} else if flag.NArg() == 1 {
		s.inName = flag.Arg(0)
	} else if flag.NArg() > 1 {
		fmt.Print("Too many arguments\n\n")
		printHelp()
		return
	}
	if lDir != "" {
		d.lineDir = lDir
	}

	// Read the content
	// Do a first pass through the content
	fReader, err := fileReader(s.inName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	processContent(fReader, &s, &d, proc)
	if err := fReader.Close(); err != nil {
		fmt.Println(err.Error())
		return
	}

	// Check code chunks and maybe abort
	d.lat = compileLattice(d.chunks)
	errs := make([]error, 0)
	if err := assertTopLevelChunksAreFilenames(d.lat); err != nil {
		errs = append(errs, err)
	}
	if err := assertNoCycles(d.lat); err != nil {
		errs = append(errs, err)
	}
	if err := assertAllChunksDefined(d.chunks, d.lat); err != nil {
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
		fmt.Printf("%s: %d: %s\n", w.fName, w.line, w.msg)
	}

	// Write out the code files
	top := topLevelChunks(d.lat)
	err = writeChunks(top, d, d.lineDir, s.inName, getWriteCloser)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Write out the markdown
	if err := writeHTML(s.inName, &d); err != nil {
		fmt.Print(err.Error())
		return
	}

}

func newDoc() doc {
	return doc{
		markdown:    make(map[string]*strings.Builder),
		chunks:      make(map[string]*chunk),
		chunkStarts: make(map[int]string),
		chunkRefs:   make(map[int]chunkRef),
		secStarts:   make(map[int]section),
	}
}
func fileReader(fName string) (io.ReadCloser, error) {
	var f *os.File
	var err error
	if fName == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(fName)
	}

	return f, err
}

func processContent(r io.Reader, s *state, d *doc, proc lineProc) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		proc(s, d, sc.Text())
	}

	if s.inChunk {
		s.warnings = append(s.warnings,
			warning{s.inName, s.lineNum,
				"Content finished but chunk not closed"})
	}
}

func proc(s *state, d *doc, line string) {
	s.lineNum++
	// Track and mark section changes
	if !s.inChunk && strings.HasPrefix(line, "#") {
		var changed bool
		s.sec, changed = s.sec.next(line)
		if changed {
			line = strings.Repeat("#", len(s.sec.nums)) +
				" <a name=\"sec" + s.sec.numsToString() + "\"></a>" +
				s.sec.toString()
			d.secStarts[s.lineNum] = s.sec
		}
	}

	// Collect lines in code chunks
	if s.inChunk && line == "```" {
		s.inChunk = false
		// Capture data for post-chunk references
		d.chunkRefs[s.lineNum] = chunkRef{s.chunkName, s.sec}

	} else if s.inChunk {
		d.chunks[s.chunkName].cont = append(
			d.chunks[s.chunkName].cont,
			chunkCont{
				file: s.inName,
				lNum: s.lineNum,
				code: line,
			})
	} else if !s.inChunk && strings.HasPrefix(line, "```") {
		s.chunkName = strings.TrimSpace(line[3:])
		if s.chunkName == "" {
			s.warnings = append(s.warnings,
				warning{s.inName, s.lineNum, "Chunk has no name"})
		}
		ch := d.chunks[s.chunkName]
		if ch == nil {
			d.chunks[s.chunkName] = &chunk{}
			ch = d.chunks[s.chunkName]
		}
		d.chunkStarts[s.lineNum] = s.chunkName
		d.chunks[s.chunkName].def = append(
			d.chunks[s.chunkName].def,
			chunkDef{
				file: s.inName,
				line: s.lineNum,
				sec:  s.sec,
			})
		s.inChunk = true
	}

	if _, okay := d.markdown[s.inName]; !okay {
		d.markdown[s.inName] = &strings.Builder{}
	}
	d.markdown[s.inName].WriteString(line + "\n")
}

func (s *section) toString() string {
	if len(s.nums) == 0 {
		return "0"
	}

	return s.numsToString() + " " + s.text
}

func (s *section) numsToString() string {
	num := ""
	for i, n := range s.nums {
		num += strconv.Itoa(n)
		if i < len(s.nums)-1 {
			num += "."
		}
	}
	return num
}

// next returns the section, and if it's changed, given a line of markdown.
func (s *section) next(line string) (section, bool) {
	re, _ := regexp.Compile("(#+)\\s+(.*)")
	find := re.FindStringSubmatch(line)
	if len(find) < 2 {
		return *s, false
	}

	oldLevel := len(s.nums)
	newLevel := len(find[1])
	nums := make([]int, newLevel)
	if oldLevel < newLevel {
		for i := 0; i < oldLevel; i++ {
			nums[i] = s.nums[i]
		}
		nums[newLevel-1] = 1
	} else {
		for i := 0; i < newLevel-1; i++ {
			nums[i] = s.nums[i]
		}
		nums[newLevel-1] = s.nums[newLevel-1] + 1
	}

	return section{nums, find[2]}, true
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

		for _, cont := range data.cont {
			refChunk := referredChunkName(cont.code)
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

func writeChunks(
	top []string,
	d doc,
	lineDir string,
	fName string,
	getWC func(string) (io.WriteCloser, error)) error {

	for _, name := range top {
		wc, err := getWC(name)
		if err != nil {
			return err
		}
		bw := bufio.NewWriter(wc)
		err = writeChunk(name, d, bw, lineDir, "", fName)
		if err != nil {
			wc.Close()
			return err
		}
		if err := bw.Flush(); err != nil {
			wc.Close()
			return err
		}
		if err := wc.Close(); err != nil {
			return err
		}
	}

	// No errors - all okay
	return nil
}

func getWriteCloser(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

func writeChunk(name string,
	d doc,
	w *bufio.Writer,
	lineDir string,
	indent string,
	fName string) error {

	chunk := *d.chunks[name]
	for _, cont := range chunk.cont {
		code := cont.code
		var err error
		if ref := referredChunkName(code); ref != "" {
			iPos := strings.Index(code, "@")
			err = writeChunk(ref, d, w, lineDir, code[0:iPos]+indent, fName)
		} else {
			lNum := cont.lNum
			indentHere := initialWS(code)
			dir := lineDirective(lineDir, indent+indentHere, fName, lNum)
			_, err = w.WriteString(dir + indent + code + "\n")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func initialWS(code string) string {
	whitespace, _ := regexp.Compile("^\\s*")
	res := whitespace.FindStringSubmatch(code)
	if len(res) == 0 {
		return ""
	}
	return res[0]
}

func lineDirective(dir string, indent string, fName string, n int) string {
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
			case 'i':
				out += indent
			case 'f':
				out += fName
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

func writeHTML(inName string, d *doc) error {
	oName := outName(inName)
	md := finalMarkdown(inName, d).String()
	output := markdown.ToHTML([]byte(md), nil, nil)
	outFile, err := os.Create(oName)
	if err != nil {
		return err
	}
	_, err = io.WriteString(outFile, string(output))
	if err != nil {
		outFile.Close()
		return err
	}
	return outFile.Close()
}

func finalMarkdown(inName string, d *doc) *strings.Builder {
	b := strings.Builder{}
	r := strings.NewReader(d.markdown[inName].String())
	sc := bufio.NewScanner(r)
	count := 0
	for sc.Scan() {
		count++
		mkup := sc.Text()
		// Amend chunk starts to include coding language
		if name, okay := d.chunkStarts[count]; okay {
			mkup = backticks(mkup)
			top := topOf(name, d.lat)
			re, _ := regexp.Compile("[-_a-zA-Z0-9]*$")
			langs := re.FindStringSubmatch(top)
			if langs != nil {
				mkup += langs[0]
			}
		}

		b.WriteString(mkup + "\n")
		// Include post-chunk reference if necessary
		if ref, ok := d.chunkRefs[count]; ok {
			str1 := addedToChunkRef(d, ref)
			b.WriteString(str1)
			str2 := usedInChunkRef(d, ref)
			b.WriteString(str2)
		}

	}
	return &b
}

// topOf takes a chunk name and returns the top-most parent name
func topOf(name string, lat lattice) string {
	for len(lat.parentsOf[name]) > 0 {
		// Get any parent of this chunk
		for par := range lat.parentsOf[name] {
			name = par
			break
		}
	}
	return name
}

// backticks gets all the backticks at the start of a string
func backticks(mkup string) string {
	out := ""
	for _, roon := range mkup {
		if roon != '`' {
			return out
		}
		out += "`"
	}
	return out
}

func addedToChunkRef(d *doc, ref chunkRef) string {
	chunk := d.chunks[ref.name]
	secs := make([]section, len(chunk.def))
	for i, def := range chunk.def {
		secs[i] = def.sec
	}

	for i, sec := range secs {
		if reflect.DeepEqual(ref.thisSec, sec) {
			secs = append(secs[:i], secs[i+1:]...)
			break
		}
	}

	if len(secs) == 0 {
		return ""
	}

	return "\nAdded to in " + sectionsAsEnglish(secs) + ".\n\n"
}

func sectionsAsEnglish(secs []section) string {
	list := ""
	for i, sec := range secs {
		list += sec.numsToString()
		if i < len(secs)-2 {
			list += ", "
		} else if i == len(secs)-2 {
			list += " and "
		}
	}

	prefix := "section "
	if len(secs) > 1 {
		prefix = "sections "
	}

	return prefix + list
}

func usedInChunkRef(d *doc, ref chunkRef) string {
	secs := make([]section, 0)

	// Get the sections
	for parName, _ := range d.lat.parentsOf[ref.name] {
		chunk := d.chunks[parName]
		for _, cont := range chunk.cont {
			if referredChunkName(cont.code) == ref.name {
				var sec section
				for _, def := range chunk.def {
					if def.line < cont.lNum {
						sec = def.sec
					}
				}
				secs = append(secs, sec)
			}
		}
	}

	if len(secs) == 0 {
		return ""
	}

	// Sort the sections
	sort.Slice(secs, func(i, j int) bool { return secs[i].less(secs[j]) })

	return "\nUsed in " + sectionsAsEnglish(secs) + ".\n\n"
}

func (s1 *section) less(s2 section) bool {
	n1, n2 := s1.nums, s2.nums
	var limit int
	if len(n1) < len(n2) {
		limit = len(n1)
	} else {
		limit = len(n2)
	}

	for i := 0; i < limit; i++ {
		switch {
		case n1[i] < n2[i]:
			return true
		case n1[i] > n2[i]:
			return false
		}
	}
	return len(n1) < len(n2)
}

func printHelp() {
	msg := `litgo [--line-dir <ldir>] <input-file>

    <input-file> can be - (or be omitted) to indicate stdin.
    <ldir> is the line directive to preceed each code line.
        Use %f for filename, %l for line number,
        %i to include indentation, %% for percent sign.
`
	fmt.Printf(msg)
}

func outName(fName string) string {
	base := filepath.Base(fName)
	ext := filepath.Ext(fName)
	if fName == "-" || base == "." {
		base = "out"
	}
	pref := base
	if ext != "" {
		pref = base[0 : len(base)-len(ext)]
	}
	return pref + ".html"
}
