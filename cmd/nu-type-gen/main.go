package main

import (
	"bufio"
	"io"
	"iter"
	"log"
	"os"
)

func linesIter(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		// reads included file paths via stdin to get around CLI max length
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				break
			}
		}
	}
}

func main() {
	gen, err := NewGenerator(linesIter(os.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	defer gen.Close()

	err = gen.Generate()
	if err != nil {
		log.Fatal(err)
	}
}
