# nu-type-alias

> A codemod for nushell that implements type aliases and generics.

> [!IMPORTANT]
> This project is *early-stage*, so it would be a *very good idea*
> to commit your codebase before running it. There is a non-zero
> chance that it may generate incorrect code or corrupt files due
> to bugs.

## Usage

`nu-type-alias` is a tool that allows you to write type
annotations for values in a "javadoc-esque" style and generate
actual type annotations which are enforceable during compile and
run-time.

Ex.

```nu
# type Entry = record<id: int, name: string>
# type Response = record<
#   time: timestamp,
#   entries: oneof<list<Entry>, nothing>,
#   first: oneof<Entry, nothing>
# >

# @type Response
let res = {}

# --- output (modified in-place)

# type Entry = record<id: int, name: string>
# type Response = record<
#   time: timestamp,
#   entries: oneof<list<Entry>, nothing>,
#   first: oneof<Entry, nothing>
# >

# @type Response
let res: record<time: timestamp, entries: oneof<list<record<id: int, name: string>>, nothing>, first: oneof<record<id: int, name: string>, nothing>> = {}
```

It also allows you to define generics for your custom type aliases
and annotate other values like parameters of closures and
commands.

### Type definitions

Types are defined within comments. Generics are supported and
newlines can be used between type arguments.

```nu
# Syntax:
# type ID<arg1, arg2, ...> = {type_expr}

# type Range<T> = record<
#   min: T,
#   max: T
# >

# type ListOfRanges<T> = list<Range<T>>
```

### Variable types

You can declare a variable's type with the syntax `@type`.

```nu
## before

# @type Range<int>
let foo = {min: 34, max: 70}

## after

# @type Range<int>
let foo: record<min: int, max: int> = {min: 34, max: 70}
```

> [!IMPORTANT]
> Any existing type annotation for the variable will be
> overwritten.

You will not be able to use these types outside the file unless
you have exported them (see [[#Importing/exporting types]]).

### Command types

You can declare the input and output as well as parameter types of
a command using:

- `@input {type_expr}`
- `@output {type_expr}`
- `@param {param} {type_expr}`

```nu
## before

# type PadCfg = record<left: bool, right: bool>
#
# @input Range<number>
# @output string
# @param pad PadCfg
def format [--pad] { ... }

## after

# type PadCfg = record<left: bool, right: bool>
#
# @input Range<number>
# @output string
# @param pad PadCfg
def format [--pad: record<left: bool, right: bool>]: record<min: number, max: number> -> string {
    ...
}
```

### Closure types

Closures in nushell currently only support parameter typing. You
can annotate closure parameters in the same way as commands, just
make sure to put them in a comment block right after the parameter
list, before any commands or expressions.

```nu
## before

{|foo|
    # you must put param types here
    #
	# @param foo Range<string>

	echo "hello"

    # you will not be able to use `@param` after any expressions
}

## after

{|foo: record<left: string, right: string>|
    # you must put param types here
    #
	# @param foo Range<string>
	echo "hello"

    # you will not be able to use `@param` after any expressions
}
```

### Importing/exporting types

Types can also be exported and imported from files using the
`@usetype "{path}.nu"` and `export` syntax.

```nu
# foo.nu

# export type File = record<
#   name: string,
#   isdir: bool,
#   children: list<string>
# >
```

```nu
# bar.nu

# @usetype "foo.nu"

## before

# @type list<foo.File>
let files = []

## after

# @type list<foo.File>
let files: list<record<name: string, isdir: bool, children: list<string>>> = []
```

To reference types from other files imported via `@usetype` you
**must** use a qualified name where the name is given by the
filename, minus `.nu` and all special characters and whitespace.

> `this-is a module.gen.nu` $\to$ `thisisamodulegen`

> [!IMPORTANT]
> Unlike the `use` keyword, you must put double quotes around the
> filename of `@usetype` otherwise a silent failure may occur!

### CLI

Run the `nu-type-alias` binary with a list of paths to nushell
files to be included separated by newline. The tool will modify
the files in all the given file paths with the

The following is a common pattern, it will include all nushell
files under the current directory.

```nu
ls **/*.nu
	| get name
	| str join "\n"
	| nu-type-alias
```

> [!NOTE]
> Not including certain files will cause those files to "cease to
> exist" in the eyes of the code mod, including those imported via
> `@usetype`. Therefore, all files that have anything to do with
> type aliases must be included.

## Building

1. Clone with submodules included.
2. Build `./cmd/nu-type-alias` with `CGO_ENABLED=1`

# nu-type-fmt

This project also contains a small helper binary that can format
nushell type expressions into a nicer format.

## Usage

The type expression (ex. `record<...>`) should be passed into
the STDIN of the command, which will output the formatted string
in STDOUT.

```
Usage of nu-type-fmt:
  -indent string
    	The indent to use. (default "  ")
  -max-subtree int
    	The maximum size of a type subtree (including the root) that may be present on a single line. (default 3)
```

Here is an example input and output (with `-max-subtree 7`).

```
record<id: oneof<nothing, int>, profile_id: oneof<nothing, int>, state: oneof<nothing, record<name: oneof<nothing, string>, desc: oneof<nothing, string>, timescale: oneof<nothing, int>, duration_cfg: oneof<nothing, record<pert: oneof<nothing, record<pes: oneof<nothing, duration>, exp: oneof<nothing, duration>, opt: oneof<nothing, duration>>>, deadline: oneof<nothing, datetime>, total_cost: oneof<nothing, int>>>, children_cfgs: list<record<desc: oneof<nothing, string>, deadline: oneof<nothing, datetime>, exp_cost: oneof<nothing, int>, children: list<record<id: oneof<nothing, int>, name: oneof<nothing, string>>>>>, prereqs: list<record<id: oneof<nothing, int>, name: oneof<nothing, string>>>, postreqs: list<record<id: oneof<nothing, int>, name: oneof<nothing, string>>>, parent: oneof<nothing, record<id: oneof<nothing, int>, name: oneof<nothing, string>>>, start: oneof<nothing, datetime>, end: oneof<nothing, datetime>>>>
```

```
record<
  id: oneof<nothing, int>
  profile_id: oneof<nothing, int>
  state: oneof<
    nothing
    record<
      name: oneof<nothing, string>
      desc: oneof<nothing, string>
      timescale: oneof<nothing, int>
      duration_cfg: oneof<
        nothing
        record<
          pert: oneof<
            nothing
            record<
              pes: oneof<nothing, duration>
              exp: oneof<nothing, duration>
              opt: oneof<nothing, duration>
            >
          >
          deadline: oneof<nothing, datetime>
          total_cost: oneof<nothing, int>
        >
      >
      children_cfgs: list<
        record<
          desc: oneof<nothing, string>
          deadline: oneof<nothing, datetime>
          exp_cost: oneof<nothing, int>
          children: list<
            record<id: oneof<nothing, int>, name: oneof<nothing, string>>
          >
        >
      >
      prereqs: list<
        record<id: oneof<nothing, int>, name: oneof<nothing, string>>
      >
      postreqs: list<
        record<id: oneof<nothing, int>, name: oneof<nothing, string>>
      >
      parent: oneof<
        nothing
        record<id: oneof<nothing, int>, name: oneof<nothing, string>>
      >
      start: oneof<nothing, datetime>
      end: oneof<nothing, datetime>
    >
  >
>
```

