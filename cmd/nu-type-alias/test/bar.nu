# @usetype "foo.api.nu"

# @type fooapi.PERT<string>
let x: record<opt: string, exp: string, pes: string> = {}

{|foo: any|
	# @param foo fooapi.PERT<string>
	echo "hello"
}

# @type closure
let x = (
	do {|bar, foo|
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
