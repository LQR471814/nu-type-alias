package tsquery

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// The given Node must be of GrammarName == "parameter"
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
func (w ParameterNode) GetLongID() (node *tree_sitter.Node) {
	w.assertNode()

	nameNode := w.ChildByFieldName("param_name")
	if nameNode != nil {
		node = nameNode
		return
	}

	paramShortFlag := w.ChildByFieldName("param_short_flag")
	paramLongFlag := w.ChildByFieldName("param_long_flag")

	switch {
	case paramShortFlag != nil:
		shortFlagID := paramShortFlag.ChildByFieldName("name")
		if shortFlagID == nil || shortFlagID.GrammarName() != "param_short_flag_identifier" {
			panic("assert failed: param_short_flag_identifier must be present under param_short_flag by key 'name'")
		}
		node = shortFlagID
	case paramLongFlag != nil:
		longFlagID := paramLongFlag.NamedChild(0)
		if longFlagID == nil || longFlagID.GrammarName() != "long_flag_identifier" {
			panic("assert failed: long_flag_identifier must be the first child of param_long_flag")
		}
		node = longFlagID
	case paramShortFlag != nil && paramLongFlag != nil:
		panic("assert: param_short_flag and param_long_flag cannot both be present under parameter at the same time")
	case paramShortFlag == nil && paramLongFlag == nil:
		panic("assert failed: either param_name, param_short_flag, or param_long_flag must be specified")
	}

	return
}

func (w ParameterNode) GetTypeNode(cursor *tree_sitter.TreeCursor) *tree_sitter.Node {
	for _, child := range w.NamedChildren(cursor) {
		if child.GrammarName() == "param_type" {
			return &child
		}
	}
	return nil
}

// GetFullIDRange gets the range of the entire ID (ex.
// `--flag(-f)`)
func (w ParameterNode) GetFullIDRange() (out ByteRange) {
	w.assertNode()
	longID := w.GetLongID()
	out = NewByteRange(longID.ByteRange())
	flagCapsule := w.ChildByFieldName("flag_capsule")
	if flagCapsule != nil {
		_, out.End = flagCapsule.ByteRange()
	}
	return
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
