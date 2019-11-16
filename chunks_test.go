package main

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// A writer that will eventually produce an error
type BadBuilder struct {
	sb     strings.Builder
	writes int
}

func (b *BadBuilder) WriteString(s string) (int, error) {
	if b.writes >= 3 {
		return 0, fmt.Errorf("BadBuilder has given up writing")
	}
	b.writes++
	return b.sb.WriteString(s)
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
	// Another gap
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
	wFact := func(name string) (io.StringWriter, error) {
		outputs[name] = &strings.Builder{}
		w := outputs[name]
		return w, nil
	}
	top := []string{"One", "Two"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1, 16},
			[]string{"Line 1.1", "@{Three}", "Line 1.3", "Line 1.4"},
			[]int{2, 3, 4, 17},
		},
		"Two": &chunk{
			[]int{7},
			[]string{"@{Three}", "Line 2.2"},
			[]int{8, 9},
		},
		"Three": &chunk{
			[]int{12},
			[]string{"Line 3.1", "Line 3.2"},
			[]int{13, 14},
		},
	}

	err := writeChunks(top, chunks, wFact)

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

	outputs := make(map[string]*BadBuilder)
	wFact := func(name string) (io.StringWriter, error) {
		outputs[name] = &BadBuilder{}
		return outputs[name], nil
	}

	top := []string{"One", "Two"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1, 16},
			[]string{"Line 1.1", "@{Three}", "Line 1.3", "Line 1.4"},
			[]int{2, 3, 4, 17},
		},
		"Two": &chunk{
			[]int{7},
			[]string{"@{Three}", "Line 2.2"},
			[]int{8, 9},
		},
		"Three": &chunk{
			[]int{12},
			[]string{"Line 3.1", "Line 3.2"},
			[]int{13, 14},
		},
	}

	err := writeChunks(top, chunks, wFact)
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
	wFact := func(n string) (io.StringWriter, error) { return &b, nil }
	top := []string{"One"}
	chunks := map[string]*chunk{
		"One": &chunk{
			[]int{1},
			[]string{"  Line 1.1", "  @{Two}", "  Line 1.3", "  @{Three}"},
			[]int{2, 3, 4, 5},
		},
		"Two": &chunk{
			[]int{8},
			[]string{"Line 2.1", "  @{Three}", "Line 2.2"},
			[]int{9, 10, 11},
		},
		"Three": &chunk{
			[]int{14},
			[]string{"Line 3.1"},
			[]int{15},
		},
	}

	err := writeChunks(top, chunks, wFact)

	if err != nil {
		t.Errorf("Should not have produced an error, but got %q",
			err.Error())
	}

	if b.String() != expected {
		t.Errorf("Expected\n%q\nbut got\n%q",
			expected, b.String())
	}
}
