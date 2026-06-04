package main

import (
	"bytes"
	"fmt"
	"nu-type-alias/cmd/nu-type-gen/tsquery"
	"nu-type-alias/internal/grammar"
	"regexp"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Annot interface {
	// Start should return the starting byte offset of the node targeted
	ByteStart() uint
}

type VarTypeAnnot struct {
	Annot grammar.VarTypeAnnot
	Var   *tree_sitter.Node
}

func (c VarTypeAnnot) ByteStart() uint {
	start, _ := c.Var.ByteRange()
	return start
}

func (c VarTypeAnnot) TargetRange() tsquery.ByteRange {
	n := c.Var.ChildByFieldName("type")
	if n != nil {
		return tsquery.NewByteRange(n.ByteRange())
	}
	id := c.Var.ChildByFieldName("var_name")
	_, idEnd := id.ByteRange()

	start := idEnd
	end := start

	value := c.Var.ChildByFieldName("value")
	if value != nil {
		valueStart, _ := value.ByteRange()
		end = valueStart - 3
	}

	return tsquery.NewByteRange(start, end)
}

type CmdTypeAnnot struct {
	Input  *grammar.InputTypeAnnotation
	Output *grammar.OutputTypeAnnotation
	Params []grammar.ParamTypeAnnotation
	Cmd    *tree_sitter.Node
}

func (c CmdTypeAnnot) ByteStart() uint {
	start, _ := c.Cmd.ByteRange()
	return start
}

func (c CmdTypeAnnot) GetParamType(param string) (out grammar.TypeExpr, ok bool) {
	for _, an := range c.Params {
		if an.ID == param {
			out = an.Type
			ok = true
			return
		}
	}
	ok = false
	return
}

func (c CmdTypeAnnot) IOTargetRange() tsquery.ByteRange {
	returnType := c.Cmd.ChildByFieldName("return_type")
	if returnType != nil {
		return tsquery.NewByteRange(returnType.ByteRange())
	}
	_, paramsEnd := c.Cmd.ChildByFieldName("parameters").ByteRange()
	bodyStart, _ := c.Cmd.ChildByFieldName("body").ByteRange()
	start := paramsEnd
	end := bodyStart - 1
	return tsquery.NewByteRange(start, end)
}

type AnnotationVisitor struct {
	code      []byte
	TypeDecls []grammar.TypeDecl
	UseDecls  []grammar.UseDecl
	CmdAnnots []CmdTypeAnnot
	VarAnnots []VarTypeAnnot
}

func NewAnnotationVisitor(code []byte) *AnnotationVisitor {
	return &AnnotationVisitor{code: code}
}

func (v *AnnotationVisitor) VisitLoneComment(span tsquery.ByteRange) {
	var out []grammar.Stmt
	err := parseStatements(v.code[span.Start:span.End], &out)
	if err != nil {
		panic(err)
	}
	for _, stmt := range out {
		switch stmt := stmt.(type) {
		case grammar.UseDecl:
			v.UseDecls = append(v.UseDecls, stmt)
		case grammar.TypeDecl:
			v.TypeDecls = append(v.TypeDecls, stmt)
		default:
			panic(fmt.Errorf("got unexpected statement %T", stmt))
		}
	}
}

func (v *AnnotationVisitor) VisitCmdComment(span tsquery.ByteRange, cmd *tree_sitter.Node) {
	var stmts []grammar.Stmt
	err := parseStatements(v.code[span.Start:span.End], &stmts)
	if err != nil {
		panic(err)
	}
	out := CmdTypeAnnot{Cmd: cmd}
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case grammar.InputTypeAnnotation:
			if out.Input != nil {
				panic(fmt.Errorf("got multiple @input type annotations"))
			}
			out.Input = &stmt
		case grammar.OutputTypeAnnotation:
			if out.Output != nil {
				panic(fmt.Errorf("got multiple @output type annotations"))
			}
			out.Output = &stmt
		case grammar.ParamTypeAnnotation:
			out.Params = append(out.Params, stmt)
		case grammar.UseDecl:
			v.UseDecls = append(v.UseDecls, stmt)
		case grammar.TypeDecl:
			v.TypeDecls = append(v.TypeDecls, stmt)
		default:
			panic(fmt.Errorf("got unexpected statement %T", stmt))
		}
	}
	v.CmdAnnots = append(v.CmdAnnots, out)
	return
}

func (v *AnnotationVisitor) VisitVarComment(span tsquery.ByteRange, variable *tree_sitter.Node) {
	var stmts []grammar.Stmt
	err := parseStatements(v.code[span.Start:span.End], &stmts)
	if err != nil {
		panic(err)
	}
	var typeAnnot *VarTypeAnnot
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case grammar.VarTypeAnnot:
			if typeAnnot != nil {
				panic(fmt.Errorf("got multiple @type annotations"))
			}
			typeAnnot = &VarTypeAnnot{
				Annot: stmt,
				Var:   variable,
			}
		case grammar.UseDecl:
			v.UseDecls = append(v.UseDecls, stmt)
		case grammar.TypeDecl:
			v.TypeDecls = append(v.TypeDecls, stmt)
		default:
			panic(fmt.Errorf("got unexpected statement %T", stmt))
		}
	}
	if typeAnnot != nil {
		v.VarAnnots = append(v.VarAnnots, *typeAnnot)
	}
}

var commentPrefix = regexp.MustCompile("(?m)^# ?")

// parseStatements parses grammar.Stmt from a block of code which may contain
// other text that are not grammar.Stmt
//
// it does this by attempting to parse starting from each token until reaching
// an error. if the resulting structure is valid, then it will record it
//
// this allows for type aliases to live alongside regular text in comment
// blocks
func parseStatements(code []byte, out *[]grammar.Stmt) (err error) {
	code = commentPrefix.ReplaceAllLiteral(code, []byte(""))

	lexDef := lexer.TextScannerLexer // default lex.Definition

	parser, err := participle.Build[grammar.Stmt](
		grammar.StmtUnion,
	)
	if err != nil {
		return
	}

	lex, err := lexDef.Lex("in", bytes.NewBuffer(code))
	if err != nil {
		return
	}

	pos := 0
	for {
		var peek *lexer.PeekingLexer
		peek, err = lexer.Upgrade(lex)
		if err != nil {
			return
		}

		var stmt *grammar.Stmt
		stmt, err = parser.ParseFromLexer(peek,
			participle.AllowTrailing(true),
		)
		if err == nil {
			if stmt == nil {
				panic("assert failed: stmt != nil")
			}
			*out = append(*out, *stmt)
		}

		// must reconstruct lexer every time after parsing (as it will make the
		// lexer return EOF otherwise)
		lex, err = lexDef.Lex("in", bytes.NewBuffer(code[pos:]))
		if err != nil {
			panic(err)
		}

		// peek.Cursor() > 0 when something is actually parsed
		for range peek.Cursor() - 1 {
			lex.Next()
		}

		var token lexer.Token
		token, err = lex.Next()
		if err != nil {
			return
		}
		if token.EOF() {
			break
		}
		pos += len(token.Value) + token.Pos.Offset
	}

	return
}
