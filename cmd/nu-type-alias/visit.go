package main

import (
	"nu-type-alias/internal/tsquery"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type nodeVisitor interface {
	VisitLoneComment(span tsquery.ByteRange)
	VisitCmdComment(span tsquery.ByteRange, cmd *tree_sitter.Node)
	VisitClosureHeader(span tsquery.ByteRange, closure *tree_sitter.Node)
	VisitVarComment(span tsquery.ByteRange, variable *tree_sitter.Node)
}

func visitClosure(node *tree_sitter.Node, cursor *tree_sitter.TreeCursor, visitor nodeVisitor) {
	if node.GrammarName() != "val_closure" {
		panic("assert failed: visitClosure must be called with node.GrammarName == val_closure")
	}
	// we capture the range of comments at the start of a closure
	// if there are none, we simply visit the rest of the children
	named := node.NamedChildren(cursor)
	if named[0].GrammarName() != "parameter_pipes" {
		panic("assert failed: first named child of closure must be 'parameter_pipes'")
	}
	var span *tsquery.ByteRange
	for _, child := range named[1:] {
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
	visitor.VisitClosureHeader(*span, node)
}

func visitComments(node *tree_sitter.Node, cursor *tree_sitter.TreeCursor, visitor nodeVisitor) {
	if node.GrammarName() == "val_closure" {
		visitClosure(node, cursor, visitor)
	}

	children := node.Children(cursor)

	var startComment *tsquery.ByteRange
	for i, child := range children {
		visitComments(&child, cursor, visitor)

		if child.GrammarName() != "comment" {
			startComment = nil
			continue
		}

		totalRange := tsquery.NewByteRange(child.ByteRange())

		if startComment == nil {
			startComment = &totalRange
		} else {
			totalRange.Start = startComment.Start
		}

		if i >= len(children)-1 {
			visitor.VisitLoneComment(totalRange)
			break
		}

		next := children[i+1]
		switch next.GrammarName() {
		case "decl_def":
			visitor.VisitCmdComment(totalRange, &next)
		case "stmt_let", "stmt_mut":
			visitor.VisitVarComment(totalRange, &next)
		}
	}
}
