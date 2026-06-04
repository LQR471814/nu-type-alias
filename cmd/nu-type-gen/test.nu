ls test/**/*.nu
	| get name
	| each { $"($in)\n" }
	| str join ""
	| go run .
