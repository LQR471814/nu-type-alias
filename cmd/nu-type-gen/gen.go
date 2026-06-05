package main

/*
#cgo CFLAGS: -std=c11 -I../../tree-sitter-nu/src
#include "../../tree-sitter-nu/src/parser.c"
#include "../../tree-sitter-nu/src/scanner.c"
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unsafe"

	"nu-type-alias/cmd/nu-type-gen/tsquery"
	"nu-type-alias/internal/grammar"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type File struct {
	tree       *tree_sitter.Tree
	treeCursor *tree_sitter.TreeCursor
	generator  *Generator

	Code []byte
	Path string
	// set of modules names -> relative file path
	UseDecls  map[string]string
	TypeDecls map[string]grammar.TypeDecl
	CmdAnnots []CmdTypeAnnot
	VarAnnots []VarTypeAnnot
}

func (f File) Close() {
	f.treeCursor.Close()
	f.tree.Close()
}

func (file File) resolveModule(mod string) (resolved File, err error) {
	relPath, modExist := file.UseDecls[mod]
	if !modExist {
		err = fmt.Errorf(
			"type module %v has not been declared",
			mod,
		)
		return
	}
	normalized := filepath.Clean(filepath.Join(filepath.Dir(file.Path), relPath))
	resolved, ok := file.generator.Files[normalized]
	if !ok {
		err = fmt.Errorf(
			"type module %v doesn't exist at filepath or hasn't been included: %v",
			mod,
			normalized,
		)
	}
	return
}

func (file File) resolveTypeID(id []string) (decl grammar.TypeDecl, err error) {
	decl, ok := file.TypeDecls[id[0]]
	if ok {
		return
	}
	if len(id) != 2 {
		err = fmt.Errorf(
			"could not resolve local type identifier: %v",
			strings.Join(id, "."),
		)
		return
	}
	mod := id[0]
	name := id[1]
	importedFile, err := file.resolveModule(mod)
	if err != nil {
		err = fmt.Errorf("module doesn't exist: %w", err)
		return
	}
	decl, ok = importedFile.TypeDecls[name]
	if !ok {
		err = fmt.Errorf(
			"imported type doesn't exist: %v.%v",
			mod,
			name,
		)
		return
	}
	return
}

func (file File) renderBuiltinType(expr grammar.TypeExpr, out io.Writer, callCtx *genericCallContext) {
	fmt.Fprint(out, expr.ID[0])
	if len(expr.Args) == 0 {
		return
	}
	fmt.Fprint(out, "<")
	for i, arg := range expr.Args {
		if i > 0 {
			fmt.Fprint(out, ", ")
		}
		if arg.Key != "" {
			fmt.Fprint(out, arg.Key)
			fmt.Fprint(out, ": ")
		}
		file.renderCanonType(arg.Value, out, callCtx)
	}
	fmt.Fprint(out, ">")
}

func (file File) getAnnotOrdered() []Annot {
	var annots []Annot
	for _, an := range file.CmdAnnots {
		annots = append(annots, Annot(an))
	}
	for _, an := range file.VarAnnots {
		annots = append(annots, Annot(an))
	}
	slices.SortFunc(annots, func(a, b Annot) int {
		if a.ByteStart() < b.ByteStart() {
			return -1
		}
		if a.ByteStart() > b.ByteStart() {
			return 1
		}
		return 0
	})
	return annots
}

// skipWriter writes a given buffer to an io.Writer but allows one to skip
// ranges of bytes while writing
type skipWriter struct {
	Out    io.Writer
	buff   []byte
	cursor uint
}

func newSkipWriter(buff []byte, out io.Writer) skipWriter {
	return skipWriter{
		buff: buff,
		Out:  out,
	}
}

// Next moves the cursor to the next location where replaceRange is the range
// of bytes to be replaced with something else (or nothing)
func (w *skipWriter) Next(skip tsquery.ByteRange) {
	w.Out.Write(w.buff[w.cursor:skip.Start])
	w.cursor = skip.End
}

// Remainder writes the remainder of the buffer to the io.Writer
func (w *skipWriter) Remainder() {
	if w.cursor >= uint(len(w.buff)) {
		return
	}
	w.Out.Write(w.buff[w.cursor:])
	w.cursor = uint(len(w.buff))
}

func (file File) genCmdParamAnnots(c CmdTypeAnnot, w *skipWriter) {
	// for debugging purpose
	cmdNode := tsquery.CommandNode{Node: c.Cmd}
	cmdName := cmdNode.GetName(file.Code)

	for _, p := range cmdNode.GetParameters(file.treeCursor) {
		param := tsquery.ParameterNode{Node: &p}
		id := param.GetLongID(file.Code)

		expr, ok := c.GetParamType(id)
		if !ok {
			panic(fmt.Errorf(
				"fail to resolve command %v parameter of name: %v",
				cmdName,
				id,
			))
		}

		w.Next(tsquery.NewByteRange(param.Node.ByteRange()))
		fmt.Fprint(w.Out, id)
		fmt.Fprint(w.Out, ": ")
		file.renderCanonType(expr, w.Out, nil)
	}
}

func (file File) genCmdIOAnnot(c CmdTypeAnnot, w *skipWriter) {
	// generate input / output type annotations
	if c.Input == nil && c.Output == nil {
		return
	}

	w.Next(c.IOTargetRange())
	fmt.Fprint(w.Out, ": ")

	if c.Input != nil {
		file.renderCanonType(c.Input.Type, w.Out, nil)
	} else {
		fmt.Fprint(w.Out, "any")
	}

	fmt.Fprint(w.Out, " -> ")

	if c.Output != nil {
		file.renderCanonType(c.Output.Type, w.Out, nil)
	} else {
		fmt.Fprint(w.Out, "any")
	}
}

func (file File) Generate(out io.Writer) {
	// here, we use skipWriter to skip over ranges of old code while writing
	w := newSkipWriter(file.Code, out)

	for _, an := range file.getAnnotOrdered() {
		switch an := an.(type) {
		case CmdTypeAnnot:
			file.genCmdParamAnnots(an, &w)
			file.genCmdIOAnnot(an, &w)
		case VarTypeAnnot:
			w.Next(an.TargetRange())
			fmt.Fprint(out, ": ")
			file.renderCanonType(an.Annot.Type, out, nil)
			fmt.Fprint(out)
		}
	}

	w.Remainder()
}

type genericCallContext struct {
	parentCtx *genericCallContext
	generics  map[string]string
}

func (ctx *genericCallContext) Resolve(id []string) (canonicalType string, ok bool) {
	if ctx == nil || len(id) != 1 {
		ok = false
		return
	}
	canonicalType, ok = ctx.generics[id[0]]
	if ok {
		return
	}
	if ctx.parentCtx == nil {
		ok = false
		return
	}
	return ctx.parentCtx.Resolve(id)
}

func (file File) newGenericCallCtx(
	decl grammar.TypeDecl,
	expr grammar.TypeExpr,
	parentCtx *genericCallContext,
) *genericCallContext {
	if len(decl.Generics) == 0 {
		return parentCtx
	}
	childCtx := &genericCallContext{
		parentCtx: parentCtx,
		generics:  make(map[string]string),
	}
	if len(expr.Args) != len(decl.Generics) {
		panic(fmt.Errorf(
			"assert failed: incorrect number of generic type arguments for type '%v' expected %v got %v",
			decl.ID,
			len(decl.Generics),
			len(expr.Args),
		))
	}
	for i, arg := range expr.Args {
		var canon strings.Builder
		file.renderCanonType(arg.Value, &canon, parentCtx)
		childCtx.generics[decl.Generics[i]] = canon.String()
	}
	return childCtx
}

// set parentCtx == nil if renderCanonType is not being called from within a
// generic type
func (file File) renderCanonType(expr grammar.TypeExpr, out io.Writer, parentCtx *genericCallContext) {
	if isBuiltinType(expr.ID) {
		file.renderBuiltinType(expr, out, parentCtx)
		return
	}
	canonicalType, ok := parentCtx.Resolve(expr.ID)
	if ok {
		fmt.Fprint(out, canonicalType)
		return
	}
	decl, err := file.resolveTypeID(expr.ID)
	if err != nil {
		panic(err)
	}
	// we resolve generic type params -> canonical types (in the context of the
	// parent ctx)
	childCtx := file.newGenericCallCtx(decl, expr, parentCtx)
	file.renderCanonType(decl.Type, out, childCtx)
}

type Generator struct {
	Files  map[string]File
	parser *tree_sitter.Parser
}

func nuLang() *tree_sitter.Language {
	return tree_sitter.NewLanguage(unsafe.Pointer(C.tree_sitter_nu()))
}

// the given files will be included
func NewGenerator(include iter.Seq[string]) (gen *Generator, err error) {
	parser := tree_sitter.NewParser()
	parser.SetLanguage(nuLang())

	gen = &Generator{
		Files:  make(map[string]File),
		parser: parser,
	}
	for path := range include {
		path, err = filepath.Abs(path)
		if err != nil {
			return
		}
		var code []byte
		code, err = os.ReadFile(path)
		if err != nil {
			return
		}
		gen.Files[path] = gen.newFile(path, code)
	}

	return
}

var modNameNotAllowed = regexp.MustCompile(`[^A-Za-z\d]`)

func deriveModuleName(filename string) string {
	if !strings.HasSuffix(filename, ".nu") {
		panic(fmt.Errorf("invalid file: '%v' must have extension .nu", filename))
	}
	return modNameNotAllowed.ReplaceAllLiteralString(filename[:len(filename)-3], "")
}

func (g *Generator) newFile(path string, code []byte) (file File) {
	tree := g.parser.Parse(code, nil)
	treeCursor := tree.Walk()

	file = File{
		Code:       code,
		Path:       path,
		tree:       tree,
		treeCursor: treeCursor,
		generator:  g,
		UseDecls:   make(map[string]string),
		TypeDecls:  make(map[string]grammar.TypeDecl),
		CmdAnnots:  nil,
		VarAnnots:  nil,
	}

	visitor := NewAnnotationVisitor(code)
	visitComments(file.tree.RootNode(), treeCursor, visitor)

	file.CmdAnnots = visitor.CmdAnnots
	file.VarAnnots = visitor.VarAnnots
	for _, decl := range visitor.TypeDecls {
		file.TypeDecls[decl.ID] = decl
	}
	for _, use := range visitor.UseDecls {
		var relPath string
		err := json.Unmarshal([]byte(use.Filename), &relPath)
		if err != nil {
			panic(err)
		}
		modname := deriveModuleName(filepath.Base(relPath))
		file.UseDecls[modname] = relPath
	}

	return file
}

func isBuiltinType(id []string) bool {
	if len(id) != 1 {
		return false
	}
	switch id[0] {
	case "nothing",
		"bool",
		"string",
		"int",
		"float",
		"number",
		"list",
		"record",
		"table",
		"duration",
		"datetime",
		"path",
		"filesize",
		"range",
		"binary",
		"closure",
		"cell-path",
		"any",
		"oneof":
		return true
	}
	return false
}

func (g *Generator) Generate() (err error) {
	for path, file := range g.Files {
		var f *os.File
		f, err = os.Create(path)
		if err != nil {
			return
		}
		file.Generate(f)
		f.Close()
	}
	return
}

func (g *Generator) Close() {
	for _, f := range g.Files {
		f.Close()
	}
	g.parser.Close()
}
