# nu-type-alias

> A codemod for nushell that implements type aliases and generics.

## Usage

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

## Importing/exporting types

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

