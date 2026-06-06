package tsquery

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type ParameterNode struct {
	*tree_sitter.Node
}

func (w ParameterNode) assertNode() {
	if w.Node.GrammarName() != "parameter" {
		panic(fmt.Errorf(
			"assert failed: TS node is not of type 'parameter' got '%v'",
			w.Node.GrammarName(),
		))
	}
}

// GetLongID gets the long name of a "parameter" tree-sitter node.
//
// paramNode is of GrammarName == "parameter"
func (w ParameterNode) GetLongID(code []byte) string {
	w.assertNode()
	nameNode := w.ChildByFieldName("param_name")
	if nameNode != nil {
		return NewByteRange(nameNode.ByteRange()).GetString(code)
	}
	paramLongFlag := w.ChildByFieldName("param_long_flag")
	if paramLongFlag == nil {
		panic("assert failed: either param_name or param_long_flag must be specified")
	}
	longFlagID := paramLongFlag.NamedChild(0)
	if longFlagID == nil || longFlagID.GrammarName() != "long_flag_identifier" {
		panic(fmt.Errorf(
			"assert failed: long_flag_identifier must be the first child of param_long_flag, got '%v'",
			longFlagID.GrammarName(),
		))
	}
	return NewByteRange(longFlagID.ByteRange()).GetString(code)
}

type CommandNode struct {
	*tree_sitter.Node
}

func (w CommandNode) assertNode() {
	if w.Node.GrammarName() != "decl_def" {
		panic("assert failed: TS node is not of type 'decl_def'")
	}
}

func (w CommandNode) GetName(code []byte) (out string) {
	w.assertNode()
	unquote := w.ChildByFieldName("unquoted_name")
	if unquote != nil {
		out = NewByteRange(unquote.ByteRange()).GetString(code)
	} else {
		quoted := w.ChildByFieldName("quoted_name")
		if quoted == nil {
			panic("assert failed: either quoted_name or unquoted_name must be set for a command")
		}
		out = NewByteRange(quoted.ByteRange()).GetString(code)
	}
	return
}

func (w CommandNode) GetParameters(cursor *tree_sitter.TreeCursor) []tree_sitter.Node {
	w.assertNode()
	paramListNode := w.ChildByFieldName("parameters")
	paramList := paramListNode.Children(cursor)
	return paramList[1 : len(paramList)-1]
}
