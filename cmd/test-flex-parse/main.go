package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"nu-type-alias/internal/grammar"
	"os"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func main() {
	parser, err := participle.Build[grammar.Stmt](
		grammar.StmtUnion,
	)
	if err != nil {
		log.Fatal(err)
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	inputBuff := bytes.NewBuffer(input)

	lexDef := lexer.TextScannerLexer // default lex.Definition
	lex, err := lexDef.Lex("in", inputBuff)
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	pos := 0
	for {
		peek, err := lexer.Upgrade(lex)
		if err != nil {
			panic(err)
		}
		stmt, err := parser.ParseFromLexer(peek,
			participle.AllowTrailing(true),
		)
		if err != nil {
			fmt.Print("\n--------\n", string(input[pos:]), "\n--------\n")
		} else {
			encoder.Encode(stmt)
		}

		// must reconstruct lexer every time after parsing (as it will make the
		// lexer return EOF otherwise)
		inputBuff = bytes.NewBuffer(input[pos:])
		lex, err = lexDef.Lex("in", inputBuff)
		if err != nil {
			panic(err)
		}

		// peek.Cursor() > 0 when something is actually parsed
		for range peek.Cursor() - 1 {
			lex.Next()
		}

		token, err := lex.Next()
		if err != nil {
			panic(err)
		}
		if token.EOF() {
			fmt.Println("EOF")
			break
		}
		pos += len(token.Value) + token.Pos.Offset
	}
}
