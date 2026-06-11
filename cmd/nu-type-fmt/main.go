package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"nu-type-alias/internal/grammar"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
)

func parse(buff []byte) (out *grammar.TypeExpr, err error) {
	parser, err := participle.Build[grammar.TypeExpr](
		grammar.StmtUnion,
	)
	if err != nil {
		return
	}
	out, err = parser.ParseBytes("in", buff)
	return
}

type CountedTypeChild struct {
	Key   string
	Value CountedTypeExpr
}

// CountedTypeExpr keeps track of the number of nodes (including the
// node itself) in the subtree
type CountedTypeExpr struct {
	Expr     grammar.TypeExpr
	Count    int
	Children []CountedTypeChild
}

// counts the number of nodes in the tree, including the root
func countNodes(expr grammar.TypeExpr) CountedTypeExpr {
	counted := CountedTypeExpr{
		Expr:  expr,
		Count: 1,
	}
	for _, arg := range expr.Args {
		child := countNodes(arg.Value)
		counted.Count += child.Count
		counted.Children = append(counted.Children, CountedTypeChild{
			Key:   arg.Key,
			Value: child,
		})
	}
	return counted
}

func formatTypeExpr(out io.Writer, counted CountedTypeExpr, field, indent string, maxSubtree, depth int) {
	expr := counted.Expr
	for range depth {
		fmt.Fprint(out, indent)
	}
	if field != "" {
		fmt.Fprintf(out, "%s: ", field)
	}
	fmt.Fprint(out, strings.Join(expr.ID, "."))
	if len(expr.Args) == 0 {
		return
	}
	fmt.Fprint(out, "<")
	if counted.Count <= maxSubtree {
		for i, arg := range counted.Children {
			if i > 0 {
				fmt.Fprint(out, ", ")
			}
			formatTypeExpr(out, arg.Value, arg.Key, indent, maxSubtree, 0)
		}
	} else {
		fmt.Fprint(out, "\n")
		for _, arg := range counted.Children {
			formatTypeExpr(out, arg.Value, arg.Key, indent, maxSubtree, depth+1)
			fmt.Fprint(out, "\n")
		}
		for range depth {
			fmt.Fprint(out, indent)
		}
	}
	fmt.Fprint(out, ">")
}

func run(indent string, maxSubtree int) (err error) {
	buff, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	expr, err := parse(buff)
	if err != nil {
		return
	}
	formatTypeExpr(os.Stdout, countNodes(*expr), "", indent, maxSubtree, 0)
	return
}

func main() {
	indent := flag.String("indent", "  ", "The indent to use.")
	maxSubtree := flag.Int("max-subtree", 3, "The maximum size of a type subtree (including the root) that may be present on a single line.")
	flag.Parse()
	err := run(*indent, *maxSubtree)
	if err != nil {
		log.Fatal(err)
	}
}
