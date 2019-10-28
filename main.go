// main.go
package main

import (
	// Imports
	"bufio"
	"fmt"
	"github.com/russross/blackfriday"
	"strings"
)

// Functions
func main() {
	input := []byte("# Hello world\n\nThis is my other literate document")
	output := blackfriday.Run(input)
	fmt.Println(string(output))
}

func processContent(c []byte, proc func(line string)) {
	r := strings.NewReader(string(c))
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		proc(sc.Text())
	}
}
