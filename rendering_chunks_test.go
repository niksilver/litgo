package main

import (
	"github.com/gomarkdown/markdown/ast"
	"io"
	"strings"
	"testing"
)

func Test_RenderChunk_CodeBlockMarkup(t *testing.T) {
	code := "import something\n" +
		"export something\n"
	data := map[string]string{
		"program.md": "# Section one\n" +
			"``` main.go\n" +
			code +
			"```\n",
	}

	s := newState()
	s.setFirstInName("program.md")
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)
	d.lat = compileLattice(d.chunks)

	expected := []string{
		`<pre><code class="language-go">import something`,
		`</code></pre>`,
	}
	cb := ast.CodeBlock{
		Leaf:     ast.Leaf{Literal: []byte(code)},
		IsFenced: true,
		Info:     []byte("go"),
	}

	w := strings.Builder{}
	renderChunk(&w, &cb, &d, "program.md")
	out := w.String()

	for _, sub := range expected {
		if !strings.Contains(out, sub) {
			t.Errorf("Expected output to contain %q but it is\n%s",
				sub, out)
		}
	}

}

func Test_RenderChunk_CodeBlockLinksChunkRefs(t *testing.T) {
	code := "10 LET A$='Hiya!'\n" +
		"  @{Later bit 1}\n" + // Should generate a link to section 1.2
		"  @{Later bit 2}\n" + // Should generate a link to other file
		"40 END\n"
	data := map[string]string{
		"main.md": "* [Part 1](part1.md)\n" +
			"* [Part 2](part2.md)\n",
		"part1.md": "# Section one\n" +
			"## Section onePone\n" +
			"``` main.go\n" +
			code +
			"```\n" +
			"## Section onePtwo\n" +
			"``` Later bit 1\n" +
			"20 PRINT A$\n" +
			"```\n",
		"part2.md": "# Section two\n" +
			"``` Later bit 2\n" +
			"30 REM That was fun\n" +
			"```\n",
	}

	s := newState()
	s.setFirstInName("main.md")
	s.book = "main.md"
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)
	d.lat = compileLattice(d.chunks)

	expected := []string{
		`10 LET A$='Hiya!'`,
		`  <a href="#section-1.2">@{Later bit 1}</a>`,
		`  <a href="part2.html#section-2">@{Later bit 2}</a>`,
		`40 END`,
	}
	cb := ast.CodeBlock{
		Leaf:     ast.Leaf{Literal: []byte(code)},
		IsFenced: true,
		Info:     []byte("go"),
	}

	w := strings.Builder{}
	renderChunk(&w, &cb, &d, "part1.md")
	out := w.String()

	for _, sub := range expected {
		if !strings.Contains(out, sub) {
			t.Errorf("Expected output to contain %q but it is\n%s",
				sub, out)
		}
	}

}

func Test_RenderChunk_EscapesHTML(t *testing.T) {
	code := "abc < def\n" +
		"zxy > ijk\n" +
		"marks & spencer\n" +
		"I said \"Hi...\n" +
		"all <> some \"& many\n"
	data := map[string]string{
		"program.md": "# Section one\n" +
			"``` main.go\n" +
			code +
			"```\n",
	}

	s := newState()
	s.setFirstInName("program.md")
	s.reader = func(fName string) (io.ReadCloser, error) {
		s.lineNum = 0
		return stringReadCloser{strings.NewReader(data[fName])}, nil
	}
	d := newDoc()

	firstPassForAll(&s, &d)
	d.lat = compileLattice(d.chunks)

	expected := []string{
		`abc &lt; def`,
		`zxy &gt; ijk`,
		`I said &quot;Hi...`,
		`all &lt;&gt; some &quot;&amp; many`,
	}
	cb := ast.CodeBlock{
		Leaf:     ast.Leaf{Literal: []byte(code)},
		IsFenced: true,
		Info:     []byte("go"),
	}

	w := strings.Builder{}
	renderChunk(&w, &cb, &d, "program.md")
	out := w.String()

	for _, sub := range expected {
		if !strings.Contains(out, sub) {
			t.Errorf("Expected output to contain %q but it is\n%s",
				sub, out)
		}
	}

}
