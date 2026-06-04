package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"nu-type-alias/internal/grammar"
	"os"

	"github.com/alecthomas/participle/v2"
)

func parse[T any](buff []byte, options ...participle.Option) (err error) {
	parser, err := participle.Build[T](options...)
	if err != nil {
		return
	}

	parsed, err := parser.ParseBytes("buff", buff)
	if err != nil {
		return
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(parsed)
	return
}

func main() {
	flag.Parse()
	mode := flag.Arg(0)

	buff, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	switch mode {
	case "input_annot":
		err = parse[grammar.InputTypeAnnotation](buff)
	case "output_annot":
		err = parse[grammar.OutputTypeAnnotation](buff)
	case "param_annot":
		err = parse[grammar.ParamTypeAnnotation](buff)
	case "var_annot":
		err = parse[grammar.VarTypeAnnot](buff)
	case "type_decl":
		err = parse[grammar.TypeDecl](buff)
	case "type_expr":
		err = parse[grammar.TypeExpr](buff)
	case "stmt":
		err = parse[grammar.Stmt](buff, grammar.StmtUnion)
	case "block":
		err = parse[grammar.Block](buff, grammar.StmtUnion)
	}

	if err != nil {
		log.Fatal(err)
	}
}
