// main.go
package main

import (
	// Imports
	"fmt"
	"github.com/russross/blackfriday"
)

// Functions
func main() {
	input := []byte("# Hello world\n\nThis is my other literate document")
	output := blackfriday.Run(input)
	fmt.Println(string(output))
}

func processContent(c []byte, proc func(line string)) {
	r := string(c)
	proc(r)
}
