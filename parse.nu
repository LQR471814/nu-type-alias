use treesitter.nu

let tree = tree-sitter parse script.nu --xml
	| from xml
	| get content.0.content.0

$tree
	| treesitter visit {
		let node = $in
		if $node.tag != comment { return null }
		$node | treesitter range of
	}
	| where $it != null
	| treesitter read ranges ./script.nu

