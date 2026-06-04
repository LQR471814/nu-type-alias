const SELF_PATH = path self
let self_dir = $SELF_PATH | path parse | get parent

def run-test [type: string]: string -> nothing {
	let input = $in
	print ""
	print $input
	print ""
	$input | go run $self_dir $type
}

"string" | run-test type_expr

"type foo = int" | run-test type_decl

"export type foo = list<record<int>>" | run-test type_decl

"type Foo = table<
	id: int,
	pert: PERT<int>,
	name: string
>" | run-test type_decl

"@input nothing" | run-test input_annot

"@output nothing" | run-test output_annot

"@param arg nothing" | run-test param_annot

"@type nothing" | run-test var_annot

"@input nothing" | run-test stmt
"@output any" | run-test stmt
"@param arg string" | run-test stmt

"export type PERT<T> = record<opt: T, pes: T, exp: T>
type Foo = table<
	id: int,
	pert: PERT<int>,
	name: string
>

@type Foo" | run-test block
