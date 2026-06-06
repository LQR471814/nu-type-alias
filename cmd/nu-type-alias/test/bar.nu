# @usetype "foo.api.nu"
# @usetype "baz.nu"

use foo.api.nu
use baz.nu

# type Test = record<foo: int>

# @type fooapi.PERTContainer<Test>
let x = []

# @type baz.Test
let y = []

{|foo|
	# @param foo fooapi.PERT<string>
	echo "hello"
}

# @type closure
let x: closure = (
	do {|bar foo|
		# @param bar int
		# @param foo fooapi.PERT<string>
		[4 6]
	}
	| do {|bar|
		# @param bar list<int>
		{|| "hello" }
	}
)

try {
	do $x
} catch {|err|
	# @param err record<msg: string>
	print $err
}
