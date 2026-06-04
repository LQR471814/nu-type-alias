# read ranges reads the text for a list of text ranges
export def "read ranges" [file: path]: table<srow: int, scol: int, erow: int, ecol: int> -> list<string> {
	# NOTE: this code will break for non-ASCII text files!
	let ranges = $in
	let all_lines = open $file | lines
	$ranges | each {|range|
		let lines = $all_lines | slice $range.srow..$range.erow
		let lines = $lines | update 0 { str substring $range.scol.. }
		let lines = $lines | update (($lines | length) - 1) { str substring ..$range.ecol }
		$lines | str join "\n"
	}
}

# range of returns the range of a node
export def "range of" []: record<tag: oneof<string, nothing>, attributes: oneof<record, nothing>, content: oneof<string, table>> -> oneof<nothing, record<srow: int, scol: int, erow: int, ecol: int>> {
	let range = try { $in.attributes | select srow scol erow ecol } catch { null }
	if $range == null {
		return null
	}
	{
		srow: ($range.srow | into int)
		scol: ($range.scol | into int)
		erow: ($range.erow | into int)
		ecol: ($range.ecol | into int)
	}
}

