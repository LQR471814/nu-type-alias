# visits each node in the given xml tree
export def "each node" [visitor: closure]: record<tag: oneof<string, nothing>, attributes: oneof<record, nothing>, content: oneof<string, table>> -> list {
	let node = $in
	let res = $node | do $visitor
	let children = if ($node.content | describe) != string {
		$node.content
			| each {|child| $child | each node $visitor }
			| flatten
	} else {
		[]
	}
	[$res] ++ $children
}

def "match type alias" []: list -> bool {
	let win = $in
	let first = $win.0?
	let second = $win.1?
	if $first.tag != comment { return false }
	match $second.tag {
		stmt_let => true
		decl_def => true
		_ => false
	}
}

export def "type alias" [visitor: closure]: record<tag: oneof<string, nothing>, attributes: oneof<record, nothing>, content: oneof<string, table>> -> list {
	let node = $in
	let results = $node.content
		| window 2
		| where ($it | match type alias)
		| each {|match| $match | do $visitor }
	$node.content
		| each {|child| $child | type alias $visitor }
}
