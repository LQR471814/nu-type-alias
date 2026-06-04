use treesitter.nu
use visitor.nu

let tree = tree-sitter parse script.nu --xml
	| from xml
	| get content.0.content.0

$tree
	| visitor each node {
		let node = $in
		if $node.tag != comment { return null }
		$node | treesitter range of
	}
	| where $it != null
	| reduce --fold [] {|it,acc|
		if ($acc | is-empty) {
			return [$it]
		}
		let last = $acc | last
		# we can assume that comments will always span the end of the line
		# (since that is how they work)
		if $last.erow == $it.srow - 1 {
			$acc | update (($acc | length) - 1) {
				{
					srow: $last.srow
					erow: $it.erow
					scol: $last.scol
					ecol: $it.ecol
				}
			}
		} else {
			$acc | append $it
		}
	}
	| inspect
	| treesitter read ranges ./script.nu
	| each {
		str replace --all --multiline --regex "^# ?" ""
	}

