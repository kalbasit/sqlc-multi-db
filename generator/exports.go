package generator

// This file exports internal functions for use in tests and by external callers.

// ExprToString converts an AST expression to its string representation.
// This is exported for testing purposes.
var ExprToString = exprToString

// IsDomainStructFunc checks if a type string represents a domain struct.
var IsDomainStructFunc = isDomainStructFunc

// ZeroValue returns the zero value expression for a given type string.
var ZeroValue = zeroValue

// ExtractBulkFor extracts the @bulk-for annotation value from a comment.
var ExtractBulkFor = extractBulkFor

// ToSingular converts a plural word to singular form.
var ToSingular = toSingular

// JoinParamsSignature joins parameters into a function signature string.
var JoinParamsSignature = joinParamsSignature

// JoinReturns joins return types into a comma-separated string.
var JoinReturns = joinReturns

// IsSlice checks if a type string represents a slice.
var IsSlice = isSlice

// FirstReturnType returns the first return type from a Returns slice.
var FirstReturnType = firstReturnType

// HasParam checks if a parameter with the given name exists.
var HasParam = hasParam

// ParamHasField checks if a parameter's struct type has a given field.
var ParamHasField = paramHasField

// HasSliceField checks if a struct has a slice field.
var HasSliceField = hasSliceField

// GetSliceField returns the first slice field of a struct.
var GetSliceField = getSliceField

// ToSnakeCase converts a CamelCase string to snake_case.
var ToSnakeCase = toSnakeCase

// Quote wraps a string in engine-appropriate quotes.
var Quote = quote

// GenerateFieldConversion generates field conversion code.
var GenerateFieldConversion = generateFieldConversion

// WrapperTemplate is the template for generating wrapper files.
const WrapperTemplate = wrapperTemplate
