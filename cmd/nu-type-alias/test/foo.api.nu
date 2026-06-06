# type PERT<T> = record<opt: T, exp: T, pes: T>

# @param scale float
# @input PERT<int>
# @output PERT<float>
def bar [--scale] {
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

