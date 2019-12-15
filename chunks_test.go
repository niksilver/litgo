package main

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// A string.Builder we can also close
type builderWriteCloser struct {
	*strings.Builder
}

func (bwc builderWriteCloser) Close() error {
	return nil
}

// A writer that will eventually produce an error
type badWriteCloser struct {
	sb     *strings.Builder
	writes int
}

func (bwc badWriteCloser) Write(p []byte) (n int, err error) {
	if bwc.writes+len(p) >= 12 {
		return 0, fmt.Errorf("Bad writer has given up writing")
	}
	bwc.writes += len(p)
	return bwc.sb.Write(p)
}

func (bwc badWriteCloser) Close() error {
	return nil
}

func TestWriteChunks_Okay(t *testing.T) {
	// Test code that looks like this (with line numbers):
	//
	// ``` One    1
	// Line 1.1   2
	// @{Three}   3
	// Line 1.3   4
	// ```
	// Gap line
	// ``` Two    7
	// @{Three}   8
	// Line 2.2   9
	// ```
	// Gap line
	// ``` Three  12
	// Line 3.1   13
	// Line 3.2   14
	// ```
	// ``` One    16
	// Line 1.4   17
	// ```        18

	oneExpected := `Line 1.1
Line 3.1
Line 3.2
Line 1.3
Line 1.4
`
	twoExpected := `Line 3.1
Line 3.2
Line 2.2
`

	outputs := make(map[string]*strings.Builder)
	wFact := func(name string) (io.WriteCloser, error) {
		outputs[name] = &strings.Builder{}
		b := outputs[name]
		return builderWriteCloser{b}, nil
	}
	top := []string{"One", "Two"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1, 16},
			[]section{},
			[]string{"Line 1.1", "@{Three}", "Line 1.3", "Line 1.4"},
			[]int{2, 3, 4, 17},
		},
		"Two": &chunk{
			[]int{7},
			[]section{},
			[]string{"@{Three}", "Line 2.2"},
			[]int{8, 9},
		},
		"Three": &chunk{
			[]int{12},
			[]section{},
			[]string{"Line 3.1", "Line 3.2"},
			[]int{13, 14},
		},
	}

	err := writeChunks(top, state{chunks: chunks}, wFact)

	if err != nil {
		t.Errorf("Should not have produced an error, but got %q",
			err.Error())
	}

	if len(outputs) != 2 {
		t.Errorf("Should have two top chunks output, but got %d: %v",
			len(outputs), reflect.ValueOf(outputs).MapKeys())
	}

	if outputs["One"] == nil {
		t.Errorf("Chunk One did not have a Builder")
	} else if outputs["One"].String() != oneExpected {
		t.Errorf("For chunk One expected\n%q\nbut got\n%q",
			oneExpected, outputs["One"].String())
	}

	if outputs["Two"] == nil {
		t.Errorf("Chunk Two did not have a Builder")
	} else if outputs["Two"].String() != twoExpected {
		t.Errorf("For chunk Two expected\n%q\nbut got\n%q",
			twoExpected, outputs["Two"].String())
	}
}

func TestWriteChunks_ErrorWriting(t *testing.T) {
	// Test code that looks like this (with line numbers):
	//
	// ``` One    1
	// Line 1.1   2
	// @{Three}   3
	// Line 1.3   4
	// ```
	// Gap line
	// ``` Two    7
	// @{Three}   8
	// Line 2.2   9
	// ```
	// Another gap
	// ``` Three  12
	// Line 3.1   13
	// Line 3.2   14
	// ```
	// ``` One    16
	// Line 1.4   17
	// ```        18

	outputs := make(map[string]*strings.Builder)
	wFact := func(name string) (io.WriteCloser, error) {
		outputs[name] = &strings.Builder{}
		w := outputs[name]
		return badWriteCloser{w, 0}, nil
	}

	top := []string{"One", "Two"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1, 16},
			[]section{},
			[]string{"Line 1.1", "@{Three}", "Line 1.3", "Line 1.4"},
			[]int{2, 3, 4, 17},
		},
		"Two": &chunk{
			[]int{7},
			[]section{},
			[]string{"@{Three}", "Line 2.2"},
			[]int{8, 9},
		},
		"Three": &chunk{
			[]int{12},
			[]section{},
			[]string{"Line 3.1", "Line 3.2"},
			[]int{13, 14},
		},
	}

	err := writeChunks(top, state{chunks: chunks}, wFact)
	if err == nil {
		t.Errorf("Should have produced an error, did not")
	}
}

func TestWriteChunks_IndentProperly(t *testing.T) {
	// Test code that looks like this (with line numbers):
	//
	// ``` One      1
	//   Line 1.1   2
	//   @{Two}     3  Indent Two by two spaces
	//   Line 1.3   4
	//   @{Three}   5  Indent Three by two spaces where here
	// ```
	// Gap line
	// ``` Two      8
	// Line 2.1     9
	//   @{Three}   10 Indent Three by two more spaces where here
	// Line 2.2     11
	// ```
	// Another gap
	// ``` Three    14
	// Line 3.1     15
	// ```

	expected := `  Line 1.1
  Line 2.1
    Line 3.1
  Line 2.2
  Line 1.3
  Line 3.1
`

	b := strings.Builder{}
	wFact := func(n string) (io.WriteCloser, error) {
		return builderWriteCloser{&b}, nil
	}
	top := []string{"One"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1},
			[]section{},
			[]string{"  Line 1.1", "  @{Two}", "  Line 1.3", "  @{Three}"},
			[]int{2, 3, 4, 5},
		},
		"Two": &chunk{
			[]int{8},
			[]section{},
			[]string{"Line 2.1", "  @{Three}", "Line 2.2"},
			[]int{9, 10, 11},
		},
		"Three": &chunk{
			[]int{14},
			[]section{},
			[]string{"Line 3.1"},
			[]int{15},
		},
	}

	err := writeChunks(top, state{chunks: chunks}, wFact)

	if err != nil {
		t.Errorf("Should not have produced an error, but got %q",
			err.Error())
	}

	if b.String() != expected {
		t.Errorf("Expected\n%q\nbut got\n%q",
			expected, b.String())
	}
}

func TestWriteChunks_OkayWithLineDirectives(t *testing.T) {
	// Test code that looks like this (with line numbers):
	//
	// ``` One    1
	// Line 1.1   2
	// @{Three}   3
	// Line 1.3   4
	// ```
	// Gap line
	// ``` Two    7
	// @{Three}   8
	// Line 2.2   9
	// ```
	// Another gap
	// ``` Three  12
	// Line 3.1   13
	// Line 3.2   14
	// ```
	// ``` One    16
	// Line 1.4   17
	// ```        18

	oneExpected := `//line test.lit:2
Line 1.1
//line test.lit:13
Line 3.1
//line test.lit:14
Line 3.2
//line test.lit:4
Line 1.3
//line test.lit:17
Line 1.4
`
	twoExpected := `//line test.lit:13
Line 3.1
//line test.lit:14
Line 3.2
//line test.lit:9
Line 2.2
`

	outputs := make(map[string]*strings.Builder)
	wFact := func(name string) (io.WriteCloser, error) {
		outputs[name] = &strings.Builder{}
		b := outputs[name]
		return builderWriteCloser{b}, nil
	}
	top := []string{"One", "Two"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1, 16},
			[]section{},
			[]string{"Line 1.1", "@{Three}", "Line 1.3", "Line 1.4"},
			[]int{2, 3, 4, 17},
		},
		"Two": &chunk{
			[]int{7},
			[]section{},
			[]string{"@{Three}", "Line 2.2"},
			[]int{8, 9},
		},
		"Three": &chunk{
			[]int{12},
			[]section{},
			[]string{"Line 3.1", "Line 3.2"},
			[]int{13, 14},
		},
	}

	s := state{chunks: chunks, fname: "test.lit", lineDir: "//line %f:%l"}
	err := writeChunks(top, s, wFact)

	if err != nil {
		t.Errorf("Should not have produced an error, but got %q",
			err.Error())
	}

	if len(outputs) != 2 {
		t.Errorf("Should have two top chunks output, but got %d: %v",
			len(outputs), reflect.ValueOf(outputs).MapKeys())
	}

	if outputs["One"] == nil {
		t.Errorf("Chunk One did not have a Builder")
	} else if outputs["One"].String() != oneExpected {
		t.Errorf("For chunk One expected\n%q\nbut got\n%q",
			oneExpected, outputs["One"].String())
	}

	if outputs["Two"] == nil {
		t.Errorf("Chunk Two did not have a Builder")
	} else if outputs["Two"].String() != twoExpected {
		t.Errorf("For chunk Two expected\n%q\nbut got\n%q",
			twoExpected, outputs["Two"].String())
	}
}

func TestWriteChunks_OkayWithLineDirectiveIndents(t *testing.T) {
	// Test code that looks like this (with line numbers):
	//
	// ``` One      1
	//   Line 1.1   2
	//   @{Two}     3
	//   Line 1.3   4
	// ```
	// Gap line
	// ``` Two      7
	// Line 2.1     8
	// ```

	expectedWith := `  //line test.lit:2
  Line 1.1
  //line test.lit:8
  Line 2.1
  //line test.lit:4
  Line 1.3
`
	expectedWithout := `//line test.lit:2
  Line 1.1
//line test.lit:8
  Line 2.1
//line test.lit:4
  Line 1.3
`

	top := []string{"One"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1},
			[]section{},
			[]string{"  Line 1.1", "  @{Two}", "  Line 1.3"},
			[]int{2, 3, 4},
		},
		"Two": &chunk{
			[]int{7},
			[]section{},
			[]string{"Line 2.1"},
			[]int{8},
		},
	}

	// Test it with an indent

	outputs1 := make(map[string]*strings.Builder)
	wFact1 := func(name string) (io.WriteCloser, error) {
		outputs1[name] = &strings.Builder{}
		b := outputs1[name]
		return builderWriteCloser{b}, nil
	}
	s1 := state{chunks: chunks, fname: "test.lit", lineDir: "%i//line %f:%l"}
	err1 := writeChunks(top, s1, wFact1)

	if err1 != nil {
		t.Errorf("With: Should not have produced an error, but got %q",
			err1.Error())
	}

	if len(outputs1) != 1 {
		t.Errorf("With: Should have one top chunk output, but got %d: %v",
			len(outputs1), reflect.ValueOf(outputs1).MapKeys())
	}

	if outputs1["One"] == nil {
		t.Errorf("With: Chunk One did not have a Builder")
	} else if outputs1["One"].String() != expectedWith {
		t.Errorf("With: For chunk One expected\n%q\nbut got\n%q",
			expectedWith, outputs1["One"].String())
	}

	// Test it without an indent

	outputs2 := make(map[string]*strings.Builder)
	wFact2 := func(name string) (io.WriteCloser, error) {
		outputs2[name] = &strings.Builder{}
		b := outputs2[name]
		return builderWriteCloser{b}, nil
	}
	s2 := state{chunks: chunks, fname: "test.lit", lineDir: "//line %f:%l"}
	err2 := writeChunks(top, s2, wFact2)

	if err2 != nil {
		t.Errorf("Without: Should not have produced an error, but got %q",
			err2.Error())
	}

	if len(outputs2) != 1 {
		t.Errorf("Without: Should have one top chunk output, but got %d: %v",
			len(outputs2), reflect.ValueOf(outputs2).MapKeys())
	}

	if outputs2["One"] == nil {
		t.Errorf("Without: Chunk One did not have a Builder")
	} else if outputs2["One"].String() != expectedWithout {
		t.Errorf("Without: For chunk One expected\n%q\nbut got\n%q",
			expectedWithout, outputs2["One"].String())
	}

}

func TestLineDirective(t *testing.T) {
	data := []struct {
		dir   string
		ind   string
		fname string
		n     int
		exp   string
	}{
		{"", "", "", 3, ""},
		{"//Two", "", "", 3, "//Two\n"},
		{"%%", "", "", 4, "%\n"},
		{"*%%", "", "", 4, "*%\n"},
		{"%%%%", "", "", 4, "%%\n"},
		{"%l", "", "", 5, "5\n"},
		{"%%l", "", "", 5, "%l\n"},
		{"a%lb", "", "", 5, "a5b\n"},
		{"a%%%lb", "", "", 6, "a%6b\n"},
		{"%l%l%%%lk", "", "", 7, "77%7k\n"},
		{"%f", "", "t.go", 7, "t.go\n"},
		{"%%f", "", "t.go", 7, "%f\n"},
		{"a%fb", "", "t.go", 8, "at.gob\n"},
		{"a%%%fb", "", "t.go", 8, "a%t.gob\n"},
		{"%f%l%%%fk", "", "t.go", 8, "t.go8%t.gok\n"},
		{"%i", "  ", "", 9, "  \n"},
		{"a%ib", "  ", "", 9, "a  b\n"},
	}

	for _, d := range data {
		act := lineDirective(d.dir, d.ind, d.fname, d.n)
		if act != d.exp {
			t.Errorf("Directive %q with indent %q in file %q at line %d, expected %q but got %q",
				d.dir, d.ind, d.fname, d.n, d.exp, act)
		}
	}
}
