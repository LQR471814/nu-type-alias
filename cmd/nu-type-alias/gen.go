package main

/*
#cgo CFLAGS: -std=c11 -I../../tree-sitter-nu/src
#include "../../tree-sitter-nu/src/parser.c"
#include "../../tree-sitter-nu/src/scanner.c"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"

	"nu-type-alias/internal/grammar"
	"nu-type-alias/internal/tsquery"

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
	Annots    []Annot
}

func (f File) Close() {
	f.treeCursor.Close()
	f.tree.Close()
}

func (file File) resolveModule(mod string) (resolved File, err error) {
	relPath, modExist := file.UseDecls[mod]
	if !modExist {
		err = fmt.Errorf(
			"type module %v has not been declared %v",
			mod,
			file.UseDecls,
		)
		return
	}
	normalized := filepath.Clean(filepath.Join(filepath.Dir(file.Path), relPath))
	resolved, ok := file.generator.Files[normalized]
	if !ok {
		err = fmt.Errorf(
			"type module %v doesn't exist at filepath or hasn't been included: %v %v",
			mod,
			normalized,
			file.UseDecls,
		)
	}
	return
}

func (file File) resolveTypeID(id []string) (decl grammar.TypeDecl, source File, err error) {
	decl, ok := file.TypeDecls[id[0]]
	if ok {
		source = file
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
	source, err = file.resolveModule(mod)
	if err != nil {
		err = fmt.Errorf("module doesn't exist: %w", err)
		return
	}
	decl, ok = source.TypeDecls[name]
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

func (file File) genParamAnnots(params []tree_sitter.Node, annots ParamTypeAnnots, w *skipWriter) {
	for _, p := range params {
		param := tsquery.ParameterNode{Node: &p}

		id := tsquery.NewByteRange(param.GetLongID().ByteRange()).GetString(file.Code)
		typeExpr, ok := annots.GetParamType(id)
		if !ok {
			continue
		}

		typeNode := param.GetTypeNode(file.treeCursor)

		// get the end of the id
		replace := param.GetFullIDRange()
		replace.Start = replace.End
		if typeNode != nil {
			_, replace.End = typeNode.ByteRange()
		}
		w.Next(replace)

		if typeExpr.ID[0] == "bool" {
			// boolean flags should not have type annotation
			continue
		}

		fmt.Fprint(w.Out, ": ")
		file.renderCanonType(typeExpr, w.Out, nil)
	}
	return
}

func (file File) genCmdParamAnnots(c CmdTypeAnnot, w *skipWriter) {
	cmdNode := tsquery.CommandNode{Node: c.Cmd}
	params := cmdNode.GetParameters(file.treeCursor)
	file.genParamAnnots(params, c.Params, w)
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

func (file File) genClosureParamAnnots(c ClosureTypeAnnot, w *skipWriter) {
	closure := tsquery.ClosureNode{Node: c.Closure}
	params := closure.GetParameters(file.treeCursor)
	file.genParamAnnots(params, c.Params, w)
}

func (file File) Generate(out io.Writer) (err error) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()

	// here, we use skipWriter to skip over ranges of old code while writing
	w := newSkipWriter(file.Code, out)

	for _, an := range file.Annots {
		switch an := an.(type) {
		case CmdTypeAnnot:
			file.genCmdParamAnnots(an, &w)
			file.genCmdIOAnnot(an, &w)
		case ClosureTypeAnnot:
			file.genClosureParamAnnots(an, &w)
		case VarTypeAnnot:
			w.Next(an.TargetRange())
			fmt.Fprint(out, ": ")
			file.renderCanonType(an.Annot.Type, out, nil)
			fmt.Fprint(out)
		}
	}

	w.Remainder()
	return
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
) (out *genericCallContext, err error) {
	if len(decl.Generics) == 0 {
		out = parentCtx
		return
	}
	childCtx := &genericCallContext{
		parentCtx: parentCtx,
		generics:  make(map[string]string),
	}
	if len(expr.Args) != len(decl.Generics) {
		err = fmt.Errorf(
			"assert failed: incorrect number of generic type arguments for type '%v' expected %v got %v",
			decl.ID,
			len(decl.Generics),
			len(expr.Args),
		)
		return
	}
	for i, arg := range expr.Args {
		var canon strings.Builder
		file.renderCanonType(arg.Value, &canon, parentCtx)
		childCtx.generics[decl.Generics[i]] = canon.String()
	}
	out = childCtx
	return
}

// set parentCtx == nil if renderCanonType is not being called from within a
// generic type
func (file File) renderCanonType(expr grammar.TypeExpr, out io.Writer, parentCtx *genericCallContext) (err error) {
	if isBuiltinType(expr.ID) {
		file.renderBuiltinType(expr, out, parentCtx)
		return
	}
	canonicalType, ok := parentCtx.Resolve(expr.ID)
	if ok {
		fmt.Fprint(out, canonicalType)
		return
	}
	decl, source, err := file.resolveTypeID(expr.ID)
	if err != nil {
		return
	}
	// we resolve generic type params -> canonical types (in the context of the
	// parent ctx)
	childCtx, err := file.newGenericCallCtx(decl, expr, parentCtx)
	if err != nil {
		return
	}
	source.renderCanonType(decl.Type, out, childCtx)
	return
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
		gen.Files[path], err = gen.newFile(path, code)
		if err != nil {
			err = fmt.Errorf("new file %v: %v", path, err)
			return
		}
	}

	return
}

var modNameNotAllowed = regexp.MustCompile(`[^A-Za-z\d]`)

func deriveModuleName(filename string) (out string, err error) {
	if !strings.HasSuffix(filename, ".nu") {
		err = fmt.Errorf("invalid file: '%v' must have extension .nu", filename)
		return
	}
	out = modNameNotAllowed.ReplaceAllLiteralString(filename[:len(filename)-3], "")
	return
}

func (g *Generator) newFile(path string, code []byte) (file File, err error) {
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
	}

	visitor := NewAnnotationVisitor(code)
	visitTS := newVisitTSNode(file.tree.RootNode(), treeCursor, visitor)
	// for debug:
	// visitTS.code = code
	visitTS.Do()
	err = visitor.Err()
	if err != nil {
		// visitor errs are largely syntax warnings, successful parses will get through
		log.Println("warn:", err)
	}

	file.Annots = visitor.AnnotsOrdered()

	for _, decl := range visitor.TypeDecls {
		file.TypeDecls[decl.ID] = decl
	}
	for _, use := range visitor.UseDecls {
		var relPath string
		err = json.Unmarshal([]byte(use.Filename), &relPath)
		if err != nil {
			return
		}
		var modname string
		modname, err = deriveModuleName(filepath.Base(relPath))
		if err != nil {
			return
		}
		file.UseDecls[modname] = relPath
	}
	return
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

func (g *Generator) genFile(path string, file File) (err error) {
	var f *os.File
	f, err = os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	err = file.Generate(f)
	if err != nil {
		err = fmt.Errorf("gen file %v: %v", path, err)
		f.Truncate(0)
		f.Seek(0, 0)
		f.Write(file.Code)
		return
	}

	return
}

func (g *Generator) Generate() (err error) {
	var errs []error
	for path, file := range g.Files {
		err = g.genFile(path, file)
		if err != nil {
			errs = append(errs, err)
		}
	}
	err = errors.Join(errs...)
	return
}

func (g *Generator) Close() {
	for _, f := range g.Files {
		f.Close()
	}
	g.parser.Close()
}
