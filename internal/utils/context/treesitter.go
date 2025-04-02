package context

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// getSymbolQueryForLanguage returns the symbol query for the given language excluding imports
func (a *DefaultAnalyzer) getSymbolQueryForLanguage(fileType string) string {
	queries := map[string]string{
		"go": `
            ; Function declarations
            ((function_declaration
                name: (identifier) @func.name) @function)

            ; Method declarations
            ((method_declaration
                receiver: (parameter_list) @method.receiver
                name: (identifier) @method.name) @method)

            ; Interface declarations
            ((type_declaration
                (type_spec
                    name: (type_identifier) @interface.name
                    type: (interface_type))) @interface)

            ; Struct declarations
            ((type_declaration
                (type_spec
                    name: (type_identifier) @struct.name
                    type: (struct_type))) @struct)

            ; Type declarations (for any other kind of type)
            ((type_declaration
                (type_spec
                    name: (type_identifier) @type.name
                    type: (_))) @type)

            ; Variable declarations
            ((var_declaration
                (var_spec
                    name: (identifier) @var.name)) @var)

            ; Constant declarations
            ((const_declaration
                (const_spec
                    name: (identifier) @const.name)) @const)
                    
            ; Individual constants in a const block
            ((const_spec
                name: (identifier) @const.name) @const)
                
            ; Constants in grouped constant declarations
            ((const_declaration
                (const_spec_list
                    (const_spec 
                        name: (identifier) @const.name))) @const)
        `,
		"javascript": `
            ; Function declarations
            ((function_declaration
                name: (identifier) @func.name) @function)

            ; Arrow functions with variable declarations
            ((variable_declarator
                name: (identifier) @func.name
                value: (arrow_function)) @function)

            ; Class declarations with methods
            ((class_declaration
                name: (identifier) @class.name) @class)

            ; Class methods
            ((method_definition
                name: (property_identifier) @method.name) @method)

            ; Object methods
            ((method_definition
                name: (property_identifier) @method.name) @method)

            ; Variables and constants
            ((lexical_declaration
                (variable_declarator
                    name: (identifier) @var.name)) @var)

            ; Object properties
            ((pair
                key: (property_identifier) @prop.name
                value: [(function_expression) (arrow_function)]) @property)
        `,
		"python": `
            ; Import statements
            ((import_statement
                name: (dotted_name) @import.path) @import)

            ((import_from_statement
                module_name: (dotted_name) @import.path) @import)

            ; Function definitions
            ((function_definition
                name: (identifier) @func.name) @function)

            ; Class definitions
            ((class_definition
                name: (identifier) @class.name) @class)

            ; Methods in classes
            ((class_definition
                body: (block
                    (function_definition
                        name: (identifier) @method.name))) @method)

            ; Class variables
            ((class_definition
                body: (block
                    (expression_statement
                        (assignment
                            left: (identifier) @cvar.name)))) @cvar)
        `,
		"rust": `
            ; Use statements (keep full path)
            ((use_declaration
                tree: (use_tree) @import.path) @import)

            ; Functions and associated functions
            ((function_item
                name: (identifier) @func.name) @function)

            ; Structs
            ((struct_item
                name: (type_identifier) @struct.name) @struct)

            ; Traits
            ((trait_item
                name: (type_identifier) @trait.name) @trait)

            ; Impl blocks
            ((impl_item
                trait: (type_identifier) @impl.trait
                type: (type_identifier) @impl.type) @impl)

            ; Methods in impls
            ((impl_item
                body: (declaration_list
                    (function_item
                        name: (identifier) @method.name))) @impl_method)
        `,
		"java": `
            ; All imports (regular, static, and wildcard)
            ((import_declaration
                (scoped_identifier) @import.path) @import)

            ((import_declaration
                "static"
                (scoped_identifier) @import.path) @import)

            ; Class declarations with generics and annotations
            ((class_declaration
                (modifiers
                    (marker_annotation
                        name: (identifier) @annotation.name))?
                name: (identifier) @class.name
                type_parameters: (type_parameters
                    (type_parameter
                        (type_identifier)))?
                interfaces: (super_interfaces
                    (type_list
                        (generic_type)?))?
                body: (class_body)) @class)

            ; Methods with annotations and generics
            ((method_declaration
                (modifiers
                    (marker_annotation
                        name: (identifier) @annotation.name)?
                    (annotation
                        name: (identifier) @annotation.name
                        arguments: (annotation_argument_list)?)?)?
                type: (generic_type
                    name: (identifier)
                    type_arguments: (type_arguments))? @method.return_type
                type: (type_identifier)? @method.return_type
                name: (identifier) @method.name) @method)

            ; Constructor declarations
            ((constructor_declaration
                (modifiers
                    (marker_annotation
                        name: (identifier) @annotation.name)?)?
                name: (identifier) @constructor.name) @constructor)

            ; Fields with generics
            ((field_declaration
                (modifiers)?
                type: (generic_type
                    name: (identifier)
                    type_arguments: (type_arguments))? @field.type
                type: (type_identifier)? @field.type
                declarator: (variable_declarator
                    name: (identifier) @field.name)) @field)

            ; Annotations (both marker and regular)
            ((marker_annotation
                name: (identifier) @annotation.name) @annotation)

            ((annotation
                name: (identifier) @annotation.name
                arguments: (annotation_argument_list
                    (element_value_pair
                        key: (identifier) @annotation.key
                        value: (_) @annotation.value)?)) @annotation)

            ; Interfaces
            ((interface_declaration
                name: (identifier) @interface.name) @interface)

            ; Enum declarations
            ((enum_declaration
                name: (identifier) @enum.name) @enum)
        `,
		"cpp": `
            ; Namespace-scoped classes and their contents
            ((namespace_definition
                name: (namespace_identifier) @class.name) @class)

            ; Template classes and structs
            ((template_declaration
                parameters: (template_parameter_list)
                declaration: [(class_specifier
                    name: (type_identifier) @class.name) (struct_specifier
                    name: (type_identifier) @class.name)]) @class)

            ; Regular classes and structs
            ((class_specifier
                name: (type_identifier) @class.name) @class)
            
            ((struct_specifier
                name: (type_identifier) @struct.name) @struct)

            ; Class methods - direct function definitions in field_declaration_list
            ((field_declaration_list
                (function_definition
                    declarator: (function_declarator
                        declarator: (field_identifier) @method.name))) @method)

            ; Class methods with return types 
            ((field_declaration_list
                (function_definition
                    type: (_)?
                    declarator: (function_declarator
                        declarator: (field_identifier) @method.name))) @method)

            ; Virtual methods
            ((function_definition
                declarator: (function_declarator
                    declarator: (field_identifier) @method.name)
                    (_)? @virtual_specifier) @method)

            ; Function definition with explicit type and name
            ((function_definition
                type: (_)
                declarator: (function_declarator
                    declarator: (field_identifier) @method.name)) @method)

            ; Method function declarations inside class bodies  
            ((function_definition
                type: (_)? 
                declarator: (function_declarator
                    declarator: [(field_identifier) (identifier)] @func.name)) @function)

            ; Global functions and template functions
            ((function_definition
                declarator: (function_declarator
                    declarator: (identifier) @func.name)) @function)

            ((template_declaration
                declaration: (function_definition
                    declarator: (function_declarator
                        declarator: (identifier) @func.name))) @function)

            ; Enum definitions
            ((enum_specifier
                name: (type_identifier) @enum.name) @enum)
                
            ; Enum class constant values (individual enum values)
            ((enumerator
                name: (identifier) @const.name) @const)
                
            ; Enum constants in enumerator list
            ((enumerator_list
                (enumerator
                    name: (identifier) @const.name)) @const)
                    
            ; Regular constants and constant expressions
            ((declaration
                type: (qualified_identifier)
                declarator: (init_declarator
                    declarator: (identifier) @const.name
                    value: [(null) (number_literal) (string_literal) (char_literal)])) @const)

            ; #define constants
            ((preproc_def
                name: (identifier) @const.name) @const)
        `,
		"ruby": `
            ; Class definitions with inheritance
            ((class
                name: (constant) @class.name
                superclass: (constant)? @class.superclass) @class)

            ; Module definitions
            ((module
                name: (constant) @module.name) @module)

            ; Method definitions with parameters
            ((method
                name: (identifier) @method.name
                parameters: (method_parameters)? @method.params) @method)

            ; Singleton methods
            ((singleton_method
                object: (_) @singleton.object
                name: (identifier) @singleton.method) @singleton)

            ; Instance variables
            ((instance_variable
                "@" @ivar.symbol
                name: (_) @ivar.name) @ivar)

            ; Class variables
            ((class_variable
                "@@" @cvar.symbol
                name: (_) @cvar.name) @cvar)
                
            ; Constants
            ((assignment
                left: (constant) @const.name) @const)
                
            ; Constants in modules/classes
            ((constant) @const.name)
        `,
		"typescript": `
            ; Class declarations
            (class_declaration name: (type_identifier) @class.name) @class

            ; Interface declarations
            (interface_declaration name: (type_identifier) @interface.name) @interface

            ; Enum declarations
            (enum_declaration name: (identifier) @enum.name) @enum

            ; Type alias declarations
            (type_alias_declaration name: (type_identifier) @type.name) @type

            ; Function declarations
            (function_declaration name: (identifier) @func.name) @function

            ; Function signatures (in interfaces/types)
            (method_signature name: (property_identifier) @func.name) @function
            (call_signature) @function ; Anonymous call signatures

            ; Arrow functions assigned to variables/constants
            (lexical_declaration
              (variable_declarator
                name: (identifier) @func.name
                value: (arrow_function))) @function

            ; Methods in classes/objects, including Getters and Setters
            (method_definition name: (property_identifier) @method.name) @method
            ; Note: We capture getters/setters as methods. Differentiation might need post-processing if required.
            ; (method_definition kind: "get" name: (property_identifier) @getter.name) @method ; 'kind' seems problematic
            ; (method_definition kind: "set" name: (property_identifier) @setter.name) @method ; 'kind' seems problematic


            ; Fields/Properties in classes
            (public_field_definition name: (property_identifier) @field.name) @field

            ; Variables declared with let/var
            (lexical_declaration kind: ["let" "var"]
              (variable_declarator name: (identifier) @var.name)) @var

            ; Constants declared with const (excluding functions/components)
            (lexical_declaration kind: "const"
              (variable_declarator
                name: (identifier) @const.name
                value: [(arrow_function)]? @func_check)) @const
              (#not-match? @const.name "^[A-Z]") ; Exclude PascalCase like components
              (#eq? @func_check "") ; Ensure it's not a function assignment
        `,
		"tsx": `
            ; Class declarations
            (class_declaration name: (type_identifier) @class.name) @class

            ; Interface declarations
            (interface_declaration name: (type_identifier) @interface.name) @interface

            ; Enum declarations
            (enum_declaration name: (identifier) @enum.name) @enum

            ; Type alias declarations
            (type_alias_declaration name: (type_identifier) @type.name) @type

            ; Function declarations (including components declared as functions)
            (function_declaration name: (identifier) @func.name) @function
            (function_declaration name: (identifier) @component.name (#match? @component.name "^[A-Z]")) @react_component

            ; Function signatures (in interfaces/types)
            (method_signature name: (property_identifier) @func.name) @function
            (call_signature) @function ; Anonymous call signatures

            ; Arrow functions assigned to variables (including components)
            (lexical_declaration
              (variable_declarator
                name: (identifier) @func.name
                value: (arrow_function))) @function
            (lexical_declaration
              (variable_declarator
                name: (identifier) @component.name
                value: (arrow_function))) @react_component (#match? @component.name "^[A-Z]")


            ; Methods in classes/objects, including Getters and Setters
            (method_definition name: (property_identifier) @method.name) @method
            ; Note: We capture getters/setters as methods. Differentiation might need post-processing if required.
            ; (method_definition kind: "get" name: (property_identifier) @getter.name) @method ; 'kind' seems problematic
            ; (method_definition kind: "set" name: (property_identifier) @setter.name) @method ; 'kind' seems problematic


            ; Fields/Properties in classes
            (public_field_definition name: (property_identifier) @field.name) @field

            ; Variables declared with let/var
            (lexical_declaration kind: ["let" "var"]
              (variable_declarator name: (identifier) @var.name)) @var

            ; Constants declared with const (excluding functions/components)
            (lexical_declaration kind: "const"
              (variable_declarator
                name: (identifier) @const.name
                value: [(arrow_function)]? @func_check)) @const
              (#not-match? @const.name "^[A-Z]") ; Exclude PascalCase like components
              (#eq? @func_check "") ; Ensure it's not a function assignment
        `,
	}
	return queries[fileType]
}

// getImportQueryForLanguage returns the import query for the given language
func (a *DefaultAnalyzer) getImportQueryForLanguage(fileType string) string {
	queries := map[string]string{
		"go": `
            (import_declaration 
                (import_spec_list 
                    (import_spec 
                        path: (interpreted_string_literal) @import.path
                        name: (_)? @import.name)))

            (import_declaration 
                (import_spec 
                    path: (interpreted_string_literal) @import.path
                    name: (_)? @import.name))
        `,
		"javascript": `
            (import_statement
                source: (string_literal) @import.path
                (import_clause
                    (named_imports
                        (import_specifier
                            name: (identifier) @import.name
                            alias: (identifier)? @import.alias))))

            (import_statement
                source: (string_literal) @import.path
                (import_clause
                    (identifier) @import.default))

            (import_statement
                source: (string_literal) @import.path
                (import_clause
                    (namespace_import
                        (identifier) @import.namespace)))

            (call_expression
                function: (identifier) @require
                arguments: (arguments
                    (string_literal) @import.path)
                (#eq? @require "require"))
        `,
		"python": `
            ((import_statement
                [(dotted_name) @import.module]) @import)

            ((import_statement
                [
                    (dotted_name) @import.module
                    (identifier) @import.alias
                ]) @import)

            ((import_from_statement
                [
                    (dotted_name) @import.from
                    (dotted_name) @import.name
                ]) @import)

            ((import_from_statement
                [
                    (dotted_name) @import.from
                    (dotted_name) @import.name
                    (identifier) @import.alias
                ]) @import)

            ((import_from_statement
                [
                    (dotted_name) @import.from
                    (aliased_import
                        [(identifier) @import.star])
                ]) @import)
        `,
		"rust": `
            ((use_declaration 
                argument: (scoped_identifier) @import.path) @import)

            ((use_declaration
                argument: (identifier) @import.path) @import)
        `,
		"java": `
            ; All imports (regular, static, and wildcard)
            ((import_declaration
                (scoped_identifier) @import.path) @import)
        `,
		"cpp": `
            ; Standard library includes
            ((preproc_include
                path: (system_lib_string) @import.path) @import)

            ; Local includes
            ((preproc_include
                path: (string_literal) @import.path) @import)
        `,
		"typescript": `
            ; Import declarations (capturing only the path string literal)
            (import_statement source: (string) @import.path) @import

            ; Dynamic imports
            (call_expression
              function: (import) @dynamic_import
              arguments: (arguments (string) @import.path)) @import

            ; require calls - ensure correct structure
            (call_expression
              function: (identifier) @require (#eq? @require "require")
              arguments: (arguments (string) @import.path)) @import
        `,
		"tsx": `
            ; Import declarations (capturing only the path string literal)
            (import_statement source: (string) @import.path) @import

            ; Dynamic imports
            (call_expression
              function: (import) @dynamic_import
              arguments: (arguments (string) @import.path)) @import

            ; require calls - ensure correct structure
            (call_expression
              function: (identifier) @require (#eq? @require "require")
              arguments: (arguments (string) @import.path)) @import
        `,
	}
	return queries[fileType]
}

// getQuery returns the query for the given query type
func (a *DefaultAnalyzer) getQuery(queryType, fileType string, lang *sitter.Language) (*sitter.Query, error) {
	if lang == nil {
		return nil, fmt.Errorf("language not supported: %s", fileType)
	}

	cacheKey := fmt.Sprintf("%s_%s", queryType, fileType)

	a.mu.RLock()
	if query, ok := a.queries[cacheKey]; ok {
		a.mu.RUnlock()
		return query, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	queryStr := ""
	switch queryType {
	case "symbol":
		queryStr = a.getSymbolQueryForLanguage(fileType)
	case "import":
		queryStr = a.getImportQueryForLanguage(fileType)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", queryType)
	}

	if queryStr == "" {
		return nil, fmt.Errorf("no query available for %s in language %s", queryType, fileType)
	}

	query, err := sitter.NewQuery([]byte(queryStr), lang)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s query for %s: %w", queryType, fileType, err)
	}

	a.queries[cacheKey] = query
	return query, nil
}
