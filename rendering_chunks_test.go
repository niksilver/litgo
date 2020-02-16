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
	renderChunk(&w, &cb)
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
	renderChunk(&w, &cb)
	out := w.String()

	for _, sub := range expected {
		if !strings.Contains(out, sub) {
			t.Errorf("Expected output to contain %q but it is\n%s",
				sub, out)
		}
	}

}
