// package grammar exposes
//
// NOTE: footguns below!
//
//   - you cannot put a symbol token and a identifier together (like "@input",
//     should be "@" "input") because the two are separate tokens!
//   - do not include " ", "\t", or "\n" as tokens because they are simply
//     dropped while tokenizing
package grammar

import "github.com/alecthomas/participle/v2"

type TypeExpr struct {
	ID   string    `@Ident`
	Args []TypeArg `("<" (@@ ","?)+ ">")?`
}

type TypeArg struct {
	Key   string   `(@Ident ":")?`
	Value TypeExpr `@@`
}

type TypeDecl struct {
	Export   bool     `(@"export")? "type"`
	ID       string   `@Ident`
	Generics []string `("<" (@Ident ","?) ">")?`
	Type     TypeExpr `"=" @@`
}

type ParamTypeAnnotation struct {
	Delim struct{} `"@" "param"`
	ID    string   `@Ident`
	Type  TypeExpr `@@`
}

type InputTypeAnnotation struct {
	Delim struct{} `"@" "input"`
	Type  TypeExpr `@@`
}

type OutputTypeAnnotation struct {
	Delim struct{} `"@" "output"`
	Type  TypeExpr `@@`
}

type VarTypeAnnot struct {
	Delim struct{} `"@" "type"`
	Type  TypeExpr `@@`
}

type Stmt interface {
	stmt()
}

func (ParamTypeAnnotation) stmt()  {}
func (InputTypeAnnotation) stmt()  {}
func (OutputTypeAnnotation) stmt() {}
func (VarTypeAnnot) stmt()         {}
func (TypeDecl) stmt()             {}

var StmtUnion = participle.Union[Stmt](
	InputTypeAnnotation{},
	OutputTypeAnnotation{},
	ParamTypeAnnotation{},
	VarTypeAnnot{},
	TypeDecl{},
)

type Block struct {
	Stmts []Stmt `@@*`
}
