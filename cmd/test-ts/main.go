package main

/*
#cgo CFLAGS: -std=c11 -I../../tree-sitter-nu/src
#include "../../tree-sitter-nu/src/parser.c"
#include "../../tree-sitter-nu/src/scanner.c"
*/
import "C"
import (
	"fmt"
	"io"
	"log"
	"os"
	"unsafe"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func nuLang() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_nu())
}

func main() {
	code, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	parser := tree_sitter.NewParser()
	defer parser.Close()

	parser.SetLanguage(tree_sitter.NewLanguage(nuLang()))

	tree := parser.Parse([]byte(code), nil)
	defer tree.Close()

	walker := tree.Walk()
	defer walker.Close()

	fmt.Println(tree.RootNode().ToSexp())
}
