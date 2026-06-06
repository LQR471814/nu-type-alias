# this is a doc comment
#
# @input <type_expr>
# @output <type_expr>
# @param arg <type_expr>
export def "command name" [arg: string]: nothing -> any {
	echo "hello"
}

command name

# export type PERT<T> = record<opt: T, pes: T, exp: T>

# type Foo = table<
# 	id: int,
# 	pert: PERT<int>,
# 	name: string
# >

# @type Foo
let x = []

mut y = 4
