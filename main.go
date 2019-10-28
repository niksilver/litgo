// main.go
package main

import (
	"fmt"
	"github.com/russross/blackfriday"
)

func main() {
	input := []byte("# Hello world\n\nThis is my second literate document")
	output := blackfriday.Run(input)
	fmt.Println(string(output))
}
