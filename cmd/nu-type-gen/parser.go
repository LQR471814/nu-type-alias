package main

import (
	"bytes"
	"fmt"
	"nu-type-alias/internal/grammar"
	"regexp"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type VarTypeAnnot struct {
	Annot grammar.VarTypeAnnot
	Var   *tree_sitter.Node
}

type CmdTypeDecl struct {
	Input  *grammar.InputTypeAnnotation
	Output *grammar.OutputTypeAnnotation
	Params []grammar.ParamTypeAnnotation
	Cmd    *tree_sitter.Node
}

type AnnotationVisitor struct {
	code          []byte
	decls         []grammar.TypeDecl
	cmds          []CmdTypeDecl
	varTypeAnnots []VarTypeAnnot
}

func NewAnnotationVisitor(code []byte) *AnnotationVisitor {
	return &AnnotationVisitor{code: code}
}

func (v *AnnotationVisitor) VisitLoneComment(span ByteRange) {
	var out []grammar.Stmt
	err := parseStatements(v.code[span.Start:span.End], &out)
	if err != nil {
		panic(err)
	}
	for _, stmt := range out {
		switch stmt := stmt.(type) {
		case grammar.TypeDecl:
			v.decls = append(v.decls, stmt)
		default:
			panic(fmt.Errorf("got unexpected statement %T", stmt))
		}
	}
}

func (v *AnnotationVisitor) VisitCmdComment(span ByteRange, cmd *tree_sitter.Node) {
	var stmts []grammar.Stmt
	err := parseStatements(v.code[span.Start:span.End], &stmts)
	if err != nil {
		panic(err)
	}
	out := CmdTypeDecl{Cmd: cmd}
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
		case grammar.TypeDecl:
			v.decls = append(v.decls, stmt)
		default:
			panic(fmt.Errorf("got unexpected statement %T", stmt))
		}
	}
	v.cmds = append(v.cmds, out)
	return
}

func (v *AnnotationVisitor) VisitVarComment(span ByteRange, variable *tree_sitter.Node) {
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
		case grammar.TypeDecl:
			v.decls = append(v.decls, stmt)
		default:
			panic(fmt.Errorf("got unexpected statement %T", stmt))
		}
	}
	if typeAnnot != nil {
		v.varTypeAnnots = append(v.varTypeAnnots, *typeAnnot)
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
