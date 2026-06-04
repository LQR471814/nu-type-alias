package main

/*
#cgo CFLAGS: -std=c11 -I../../tree-sitter-nu/src
#include "../../tree-sitter-nu/src/parser.c"
#include "../../tree-sitter-nu/src/scanner.c"
*/
import "C"
import (
	"fmt"
	"unsafe"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const code = `
# this is a doc comment
#
# @input <type_expr>
# @output <type_expr>
# @param arg <type_expr>
export def "command name" [arg: string]: nothing -> any {
	echo "hello"
}

command name

# export type PERT<T> = record<opt: T, pes: T, exp: T>

# type Foo = table<
# 	id: int,
# 	pert: PERT<int>,
# 	name: string
# >

# @type Foo
let x = []

mut y = 4
`

func nuLang() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_nu())
}

func main() {
	parser := tree_sitter.NewParser()
	defer parser.Close()

	parser.SetLanguage(tree_sitter.NewLanguage(nuLang()))

	tree := parser.Parse([]byte(code), nil)
	defer tree.Close()

	walker := tree.Walk()
	defer walker.Close()

	fmt.Println(tree.RootNode().ToSexp())
}
