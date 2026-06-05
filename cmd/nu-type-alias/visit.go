package main

import (
	"nu-type-alias/internal/tsquery"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type nodeVisitor interface {
	VisitLoneComment(span tsquery.ByteRange)
	VisitCmdComment(span tsquery.ByteRange, cmd *tree_sitter.Node)
	VisitVarComment(span tsquery.ByteRange, variable *tree_sitter.Node)
}

func visitComments(node *tree_sitter.Node, cursor *tree_sitter.TreeCursor, visitor nodeVisitor) {
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
