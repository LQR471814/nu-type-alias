package main

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

type ByteRange struct {
	Start uint
	End   uint
}

func NewByteRange(start, end uint) ByteRange {
	return ByteRange{
		Start: start,
		End:   end,
	}
}

func (r ByteRange) GetString(buff []byte) string {
	return string(buff[r.Start:r.End])
}

type nodeVisitor interface {
	VisitLoneComment(span ByteRange)
	VisitCmdComment(span ByteRange, cmd *tree_sitter.Node)
	VisitVarComment(span ByteRange, variable *tree_sitter.Node)
}

func visitComments(node *tree_sitter.Node, cursor *tree_sitter.TreeCursor, visitor nodeVisitor) {
	children := node.Children(cursor)

	var startComment *ByteRange
	for i, child := range children {
		visitComments(&child, cursor, visitor)

		if child.GrammarName() != "comment" {
			startComment = nil
			continue
		}

		totalRange := NewByteRange(child.ByteRange())

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
			visitor.VisitCmdComment(totalRange, next)
		case "stmt_let", "stmt_mut":
			visitor.VisitVarComment(totalRange, next)
		}
	}
}
