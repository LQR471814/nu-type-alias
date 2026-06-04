const SELF_PATH = path self
let self_dir = $SELF_PATH | path parse | get parent

def run-test []: string -> nothing {
	let input = $in
	print ""
	print $input
	print ""
	$input | go run $self_dir
}

"this is a doc comment !

@input nothing
@output any
@param arg string

this is another comment" | run-test
