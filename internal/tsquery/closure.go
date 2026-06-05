package tsquery

import (
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type ClosureNode struct {
	*tree_sitter.Node
}

func (w ClosureNode) assertNode() {
	if w.Node.GrammarName() != "val_closure" {
		panic("assert failed: TS node is not of type 'val_closure'")
	}
}

func (w ClosureNode) GetParameters(cursor *tree_sitter.TreeCursor) []tree_sitter.Node {
	w.assertNode()
	paramListNode := w.ChildByFieldName("parameters")
	paramList := paramListNode.Children(cursor)
	return paramList[1 : len(paramList)-1]
}
