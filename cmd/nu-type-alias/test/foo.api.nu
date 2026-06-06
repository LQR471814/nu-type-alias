# type PERT<T> = record<opt: T, exp: T, pes: T>

# @input PERT<int>
# @output PERT<float>
# @param scale float
# @param f bool
# @param foo record
def bar [--scale(-s) -f --foo] {
	$in
}

# type Container<Elem> = table<
# 	label: string
# 	lists: list<Elem>
# 	first: oneof<Elem, nothing>
# >

# @type Container<PERT<int>>
let container = [[label lists first];
	[hello [] null]
]

# type PERTContainer<T> = Container<PERT<T>>

