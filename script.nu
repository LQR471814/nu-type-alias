# this is a doc comment
#
# @input nothing
# @output any
# @param arg string
export def "command name" [arg: string,    name: string, --short: int, --flag(-f): int]: nothing -> any {
	echo "hello"
}

def cmd [foo bar:int]   {

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

