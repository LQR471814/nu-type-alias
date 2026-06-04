package main

/*
#cgo CFLAGS: -std=c11 -I../../tree-sitter-nu/src
#include "../../tree-sitter-nu/src/parser.c"
#include "../../tree-sitter-nu/src/scanner.c"
*/
import "C"
import (
	"unsafe"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func nuLang() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_nu())
}

func main() {
	parser := tree_sitter.NewParser()
	defer parser.Close()

	parser.SetLanguage(tree_sitter.NewLanguage(nuLang()))

	tree := parser.Parse([]byte(code), nil)
	defer tree.Close()

	cursor := tree.Walk()
	defer cursor.Close()

	visitor := NewAnnotationVisitor([]byte(code))
	visitComments(tree.RootNode(), cursor, visitor)
}

const code = `
# this is a doc comment
#
# @input nothing
# @output any
# @param arg string
#
# this is another comment
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
