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
	book    string // Name of the top level book file, or empty if none.
	inName  string // Name of file being processed, relative to working dir
	outName string // Name of final file to write to
	// Name of all input files, including the first, relative to working dir
	inNames   []string
	lineNum   int                        // Current line number
	chunkName string                     // Name of current chunk
	inChunk   bool                       // If we're currently reading a chunk
	warnings  []warning                  // Warnings we're collecting
	sec       section                    // Current section being read
	proc      func(*state, *doc, string) // Function for processing a line
	// Function for reading a named content source (e.g. a file)
	reader func(fName string) (io.ReadCloser, error)
}

type doc struct {
	// Markdown after the initial read, per input file
	markdown map[string]*strings.Builder
	chunks   map[string]*chunk // All the chunks found so far
	// Lines where a named chunk starts, per input file
	chunkStarts map[string]map[int]string
	// Lines where other chunks are called in, per input file
	chunkRefs map[string]map[int]chunkRef
	lat       lattice // A lattice of chunk parent/child relationships
	// Lines where a section starts, per input file
	secStarts map[string]map[int]section
	// Map of normalised input file names to output names
	outNames map[string]string
	// Config
	lineDir   string // The string pattern for line directives
	docOutDir string // Output directory for the translated markdown
	// Function for opening a file to write to and close
	writeCloser func(string) (io.WriteCloser, error)
}

type warning struct {
	fName string
	line  int
	msg   string
}

type section struct {
	inName string
	nums   []int
	text   string
}

type chunk struct {
	def  []chunkDef  // Each place where the chunk is defined
	cont []chunkCont // Each line of code
}

// Where the chunk is defined: input file name, line number, section
type chunkDef struct {
	inName string
	line   int
	sec    section
}

// A line of chunk content: input file name, line number, and the code line itself
type chunkCont struct {
	inName string
	lNum   int
	code   string
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

var book bool
var lDir string
var docOutDir string

// Functions

func init() {
	// Flag initialisation
	flag.BoolVar(&book, "book", false, "If the input file is a book")
	flag.StringVar(&lDir, "line-dir", "", "Pattern for line directives")
	flag.StringVar(&docOutDir, "doc-out-dir", "", "Directory for documentation output")

}

func main() {
	// Set up the initial state
	s := newState()
	d := newDoc()

	// Update the structs according to the command line
	flag.Parse()
	if flag.NArg() == 0 {
		s.setFirstInName("-")
	} else if flag.NArg() == 1 {
		s.setFirstInName(flag.Arg(0))
	} else if flag.NArg() > 1 {
		fmt.Print("Too many arguments\n\n")
		printHelp()
		return
	}
	if book {
		s.book = s.inName
	}
	d.lineDir = lDir
	if docOutDir == "" {
		docOutDir = filepath.Dir(docOutDir)
	}
	d.docOutDir = docOutDir

	// Read the content
	// Do a first pass through all the content
	if err := firstPassForAll(&s, &d); err != nil {
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
	err := d.writeChunks(top, d.lineDir, s.inName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Write out the markdown as HTML
	if err := writeAllMarkdown(s.inNames, &d); err != nil {
		fmt.Print(err.Error())
		return
	}

}

func newState() state {
	return state{
		proc:   proc,
		reader: fileReader,
	}
}

func newDoc() doc {
	return doc{
		markdown:    make(map[string]*strings.Builder),
		chunks:      make(map[string]*chunk),
		chunkStarts: make(map[string]map[int]string),
		chunkRefs:   make(map[string]map[int]chunkRef),
		secStarts:   make(map[string]map[int]section),
		outNames:    make(map[string]string),
		writeCloser: getWriteCloser,
	}
}

func (s *state) setInName(name string) *state {
	s.inName = name
	s.sec.inName = name
	return s
}

func (s *state) setFirstInName(name string) *state {
	s.inName = name
	s.sec.inName = name
	s.inNames = []string{name}
	return s
}

func firstPassForAll(s *state, d *doc) error {
	for i := 0; i < len(s.inNames); i++ {
		inName := s.inNames[i]
		s.setInName(inName)
		if i == 0 {
			base := simpleOutName(filepath.Base(inName))
			d.outNames[inName] = filepath.Join(d.docOutDir, base)
		}
		if err := firstPass(s, d); err != nil {
			return err
		}
		s.book = ""
	}
	return nil
}

func firstPass(s *state, d *doc) error {
	fReader, err := s.reader(s.inName)
	if err != nil {
		return err
	}
	processContent(fReader, s, d)
	if err := fReader.Close(); err != nil {
		return err
	}
	if s.inChunk {
		return fmt.Errorf("File %s ended while in chunk", s.inName)
	}
	return nil
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

func processContent(r io.Reader, s *state, d *doc) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		s.proc(s, d, sc.Text())
	}

	if s.inChunk {
		s.warnings = append(s.warnings,
			warning{s.inName, s.lineNum,
				"Content finished but chunk not closed"})
	}
}

func proc(s *state, d *doc, line string) {
	s.lineNum++
	// Track chapter files to read
	foundInName := markdownLink(line)
	if s.book != "" && !s.inChunk && foundInName != "" {
		currDir := filepath.Dir(s.inName)
		normInName := filepath.Join(currDir, foundInName)
		s.inNames = append(s.inNames, normInName)
		// Update map of input to output names
		d.outNames[normInName] = chapterOutName(d.docOutDir, foundInName)

	}

	// Track and mark section changes
	if s.lineNum == 1 {
		d.addSectionStart(s.inName, s.lineNum, s.sec)
	}
	if !s.inChunk && strings.HasPrefix(line, "#") {
		var changed bool
		s.sec, changed = s.sec.next(line)
		if changed {
			d.addSectionStart(s.inName, s.lineNum, s.sec)
		}
	}

	// Collect lines in code chunks
	inChunkChanged, newChunkName := chunkChanged(&s.inChunk, line)
	if !s.inChunk && inChunkChanged {
		// Capture data for post-chunk references
		if _, okay := d.chunkRefs[s.inName]; !okay {
			d.chunkRefs[s.inName] = make(map[int]chunkRef)
		}
		d.chunkRefs[s.inName][s.lineNum] = chunkRef{s.chunkName, s.sec}

	} else if s.inChunk && !inChunkChanged {
		d.chunks[s.chunkName].cont = append(
			d.chunks[s.chunkName].cont,
			chunkCont{
				inName: s.inName,
				lNum:   s.lineNum,
				code:   line,
			})
	} else if s.inChunk && inChunkChanged {
		s.chunkName = newChunkName
		if s.chunkName == "" {
			s.warnings = append(s.warnings,
				warning{s.inName, s.lineNum, "Chunk has no name"})
		}
		ch := d.chunks[s.chunkName]
		if ch == nil {
			d.chunks[s.chunkName] = &chunk{}
			ch = d.chunks[s.chunkName]
		}
		if _, okay := d.chunkStarts[s.inName]; !okay {
			d.chunkStarts[s.inName] = make(map[int]string)
		}
		d.chunkStarts[s.inName][s.lineNum] = s.chunkName
		d.chunks[s.chunkName].def = append(
			d.chunks[s.chunkName].def,
			chunkDef{
				inName: s.inName,
				line:   s.lineNum,
				sec:    s.sec,
			})
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
	if len(s.nums) == 0 {
		return "0"
	}

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

	return section{s.inName, nums, find[2]}, true
}

func (d *doc) addSectionStart(inName string, lineNum int, sec section) {
	if _, okay := d.secStarts[inName]; !okay {
		d.secStarts[inName] = make(map[int]section)
	}
	d.secStarts[inName][lineNum] = sec
}

func markdownLink(line string) string {
	titleRE := `(\s+"[^"]*")?`
	re, _ := regexp.Compile("\\]\\(([^)#]+\\.md(#[-A-Za-z0-9_.]*)?)" + titleRE + "\\)")
	s := re.FindStringSubmatch(line)
	if len(s) == 0 {
		return ""
	}
	return s[1]
}

func chapterOutName(docOutDir string, foundInName string) string {
	return simpleOutName(filepath.Join(docOutDir, foundInName))
}

// chunkChanged sees if we're entering or leaving a chunk and updates
// `inChunk` as needed.
func chunkChanged(inChunk *bool, line string) (changed bool, newName string) {
	if *inChunk && line == "```" {
		*inChunk = false
		return true, ""
	}
	if !*inChunk && strings.HasPrefix(line, "```") {
		*inChunk = true
		return true, strings.TrimSpace(line[3:])
	}
	return false, ""
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

func (d *doc) writeChunks(
	top []string,
	lineDir string,
	fName string) error {

	for _, name := range top {
		wc, err := d.writeCloser(name)
		if err != nil {
			return err
		}
		bw := bufio.NewWriter(wc)
		err = d.writeChunk(name, bw, lineDir, "", fName)
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

func (d *doc) writeChunk(name string,
	w *bufio.Writer,
	lineDir string,
	indent string,
	fName string) error {

	chunk := d.chunks[name]
	for _, cont := range chunk.cont {
		code := cont.code
		var err error
		if ref := referredChunkName(code); ref != "" {
			iPos := strings.Index(code, "@")
			err = d.writeChunk(ref, w, lineDir, code[0:iPos]+indent, fName)
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

func writeAllMarkdown(inNames []string, d *doc) error {
	for _, inName := range inNames {
		if err := writeHTML(inName, d.outNames[inName], d); err != nil {
			return err
		}
	}
	return nil
}

func writeHTML(inName string, outName string, d *doc) error {
	md := finalMarkdown(inName, d).String()
	output := markdown.ToHTML([]byte(md), nil, nil)
	outFile, err := d.writeCloser(outName)
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
	lineNum := 0
	inChunk := false
	for sc.Scan() {
		lineNum++
		mdown := sc.Text()
		chunkChanged(&inChunk, mdown)
		// Re-link chapters and the book
		foundInName := markdownLink(mdown)
		normFoundInName := filepath.Clean(filepath.Join(filepath.Dir(inName), foundInName))
		if !inChunk && foundInName != "" && isInName(d, normFoundInName) {
			idx := strings.Index(mdown, foundInName)
			mdown = mdown[0:idx] + simpleOutName(foundInName) + mdown[idx+len(foundInName):]
		}

		// Amend section heading
		if sec, okay := d.secStarts[inName][lineNum]; okay {
			if strings.HasPrefix(mdown, "#") {
				mdown = strings.Repeat("#", len(sec.nums)) +
					" <a name=\"" + sec.anchor() + "\"></a>" +
					sec.toString()
			} else if lineNum == 1 {
				mdown = "<a name=\"" + sec.anchor() + "\"></a>\n" + mdown
			}
		}

		// Insert chunk name before start of chunk
		if name, okay := d.chunkStarts[inName][lineNum]; okay {
			b.WriteString(name + "\n\n")
		}

		// Amend chunk starts to include coding language
		if name, okay := d.chunkStarts[inName][lineNum]; okay {
			mdown = backticks(mdown)
			top := topOf(name, d.lat)
			re, _ := regexp.Compile("[-_a-zA-Z0-9]*$")
			langs := re.FindStringSubmatch(top)
			if langs != nil {
				mdown += langs[0]
			}
		}

		b.WriteString(mdown + "\n")
		// Include post-chunk reference if necessary
		if ref, ok := d.chunkRefs[inName][lineNum]; ok {
			str1 := addedToChunkRef(inName, d, ref)
			b.WriteString(str1)
			str2 := usedInChunkRef(inName, d, ref)
			b.WriteString(str2)
		}

	}
	return &b
}

func isInName(d *doc, link string) bool {
	for inName, _ := range d.markdown {
		if inName == link {
			return true
		}
	}
	return false
}

func (s *section) anchor() string {
	return "section-" + s.numsToString()
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
func backticks(mdown string) string {
	out := ""
	for _, roon := range mdown {
		if roon != '`' {
			return out
		}
		out += "`"
	}
	return out
}

func addedToChunkRef(inName string, d *doc, ref chunkRef) string {
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

	return "\nAdded to in " + sectionsAsEnglish(inName, secs) + ".\n\n"
}

func sectionsAsEnglish(inName string, secs []section) string {
	list := ""
	for i, sec := range secs {
		list += sec.link(inName)
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

func (s *section) link(hereInName string) string {
	fName, err := filepath.Rel(filepath.Dir(hereInName), s.inName)
	if err != nil {
		fName = "!!!" + err.Error() + "!!!"
	}
	return "[" + s.numsToString() + "](" + fName + "#" + s.anchor() + ")"
}

func usedInChunkRef(inName string, d *doc, ref chunkRef) string {
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

	return "\nUsed in " + sectionsAsEnglish(inName, secs) + ".\n\n"
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
	msg := `litgo [--book[=true|false]] [--line-dir <ldir>]
    [-doc-out-dir <dir>] <input-file>

    <input-file> can be - (or be omitted) to indicate stdin.

    --book[=true|false]
        Says if the input file is a book, in which case links
        to .md files are followed for that file.
    --line-dir <ldir>
        <ldir> is the line directive to preceed each code line.
        Use %f for filename, %l for line number,
        %i to include indentation, %% for percent sign.
    --doc-out-dir <dir>
        Output directory for the literate documentation. Default is
        the directory of the input file.
`
	fmt.Printf(msg)
}

func simpleOutName(fName string) string {
	if fName == "" || fName == "-" || fName == "." {
		fName = "out"
	}
	ext := filepath.Ext(fName)
	if ext != "" {
		fName = fName[0 : len(fName)-len(ext)]
	}
	return fName + ".html"
}
