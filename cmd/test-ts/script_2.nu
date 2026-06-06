export def "form done cmd" []: list<record<id: string, display_name: string, desc: string, group: string, type: oneof<record<type: string, fields: list<record<key: string, value: any>>, positional: list<any>>, record<type: string, positional: list<any>>, record<type: string, fields: list<record<key: string, value: any>>>, record<type: string>>, display_value: closure, ops: record<read: bool, write: bool, validate: oneof<closure, nothing>, >>> -> record<desc: string, group: string, aliases: list<string>, closure: record<name: string, params: list<record<key: string, value: oneof<record<type: string, fields: list<record<key: string, value: any>>, positional: list<any>>, record<type: string, positional: list<any>>, record<type: string, fields: list<record<key: string, value: any>>>, record<type: string>>>>, body: string, in: oneof<record<type: string, fields: list<record<key: string, value: any>>, positional: list<any>>, record<type: string, positional: list<any>>, record<type: string, fields: list<record<key: string, value: any>>>, record<type: string>>, out: oneof<record<type: string, fields: list<record<key: string, value: any>>, positional: list<any>>, record<type: string, positional: list<any>>, record<type: string, fields: list<record<key: string, value: any>>>, record<type: string>>, env: bool, export: bool>> {
	# @type list<types.Field>
	let fields: list<record<id: string, display_name: string, desc: string, group: string, type: oneof<record<type: string, fields: list<record<key: string, value: any>>, positional: list<any>>, record<type: string, positional: list<any>>, record<type: string, fields: list<record<key: string, value: any>>>, record<type: string>>, display_value: closure, ops: record<read: bool, write: bool, validate: oneof<closure, nothing>, >>> = $in

	let validation = $fields
		| each {|field|
			# @type types.Field
			let field: record<id: string, display_name: string, desc: string, group: string, type: oneof<record<type: string, fields: list<record<key: string, value: any>>, positional: list<any>>, record<type: string, positional: list<any>>, record<type: string, fields: list<record<key: string, value: any>>>, record<type: string>>, display_value: closure, ops: record<read: bool, write: bool, validate: oneof<closure, nothing>, >> = $field
			$"let err = (field state access) | do ($field.error | to nuon --serialize)
if $err != null {
	error make $err
}"
		}
		| str join "\n"

	let output = $fields
		| each {|field|
			# @type types.Field
			let field: record<id: string, display_name: string, desc: string, group: string, type: oneof<record<type: string, fields: list<record<key: string, value: any>>, positional: list<any>>, record<type: string, positional: list<any>>, record<type: string, fields: list<record<key: string, value: any>>>, record<type: string>>, display_value: closure, ops: record<read: bool, write: bool, validate: oneof<closure, nothing>, >> = $field
			$"\t'($field.id)': (field state access)"
		}
		| str join "\n"

	let body = $"($validation)
{
($output)
} | util save form output
exit"
	{
		desc: "Validate and submit form."
		group: control
		aliases: [d]
		closure: {
			name: done
			params: []
			body: $body
			in: {type: "nothing"}
			out: {type: "nothing"}
			env: true
			export: false
		}
	}
}

