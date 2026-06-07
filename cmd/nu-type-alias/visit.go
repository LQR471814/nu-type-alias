package main

import (
	"fmt"
	"nu-type-alias/internal/tsquery"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type nodeByNameVisitor interface {
	VisitNamedNode(node *tree_sitter.Node)
}

func visitNodeByName(
	node *tree_sitter.Node,
	cursor *tree_sitter.TreeCursor,
	name string,
	visitor nodeByNameVisitor,
) {
	if node.GrammarName() == name {
		visitor.VisitNamedNode(node)
		return
	}
	for _, child := range node.NamedChildren(cursor) {
		visitNodeByName(&child, cursor, name, visitor)
	}
}

type commentBlockVisitor interface {
	// comments in the block span children of index [start, end)
	VisitBlock(parent *tree_sitter.Node, start, end int)
}

func visitCommentBlocks(
	node *tree_sitter.Node,
	cursor *tree_sitter.TreeCursor,
	visitor commentBlockVisitor,
) {
	var prevComment *int
	children := node.NamedChildren(cursor)
	for i, child := range children {
		visitCommentBlocks(&child, cursor, visitor)
		if child.GrammarName() == "comment" {
			if prevComment == nil {
				prevComment = &i
			}
			continue
		}
		if prevComment != nil {
			visitor.VisitBlock(node, *prevComment, i)
			prevComment = nil
		}
	}
	if prevComment != nil {
		visitor.VisitBlock(node, *prevComment, len(children))
	}
}

type tsNodeVisitor interface {
	VisitLoneComment(span tsquery.ByteRange)
	VisitCmdComment(span tsquery.ByteRange, cmd *tree_sitter.Node)
	VisitClosureHeader(span tsquery.ByteRange, closure *tree_sitter.Node)
	VisitVarComment(span tsquery.ByteRange, variable *tree_sitter.Node)
}

type visitTSNode struct {
	root    *tree_sitter.Node
	cursor  *tree_sitter.TreeCursor
	visitor tsNodeVisitor
	code    []byte
}

func newVisitTSNode(
	root *tree_sitter.Node,
	cursor *tree_sitter.TreeCursor,
	visitor tsNodeVisitor,
) visitTSNode {
	return visitTSNode{
		root:    root,
		cursor:  cursor,
		visitor: visitor,
	}
}

func (b visitTSNode) Do() {
	visitCommentBlocks(b.root, b.cursor, b)
	visitNodeByName(b.root, b.cursor, "val_closure", b)
}

func (b visitTSNode) VisitBlock(parent *tree_sitter.Node, start, end int) {
	children := parent.NamedChildren(b.cursor)
	startNode := children[start]
	endNode := children[end-1]

	rng := tsquery.NewByteRange(0, 0)
	rng.Start, _ = startNode.ByteRange()
	_, rng.End = endNode.ByteRange()

	if b.code != nil {
		fmt.Println(string(b.code[rng.Start:rng.End]))
	}

	// if no next node
	if end >= len(children) {
		b.visitor.VisitLoneComment(rng)
		return
	}

	nextNode := children[end]
	switch nextNode.GrammarName() {
	case "decl_def":
		b.visitor.VisitCmdComment(rng, &nextNode)
	case "stmt_let", "stmt_mut":
		b.visitor.VisitVarComment(rng, &nextNode)
	default:
		b.visitor.VisitLoneComment(rng)
	}
}

func (b visitTSNode) VisitNamedNode(node *tree_sitter.Node) {
	if node.GrammarName() != "val_closure" {
		panic("assert failed: visitClosure must be called with node.GrammarName == val_closure")
	}
	children := node.NamedChildren(b.cursor)
	if len(children) > 0 && children[0].GrammarName() == "parameter_pipes" {
		children = children[1:]
	}
	var span *tsquery.ByteRange
	for _, child := range children {
		if child.GrammarName() != "comment" {
			break
		}
		if span == nil {
			rng := tsquery.NewByteRange(child.ByteRange())
			span = &rng
			continue
		}
		_, end := child.ByteRange()
		span.End = end
	}
	if span == nil {
		return
	}
	b.visitor.VisitClosureHeader(*span, node)
}
