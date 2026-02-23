package generator

import "go/ast"

// This file exports internal functions for use in tests and by external callers.

// ExprToString converts an AST expression to its string representation.
func ExprToString(expr ast.Expr) string { return exprToString(expr) }

// IsDomainStructFunc checks if a type string represents a domain struct.
func IsDomainStructFunc(t string) bool { return isDomainStructFunc(t) }

// ZeroValue returns the zero value expression for a given type string.
func ZeroValue(t string) string { return zeroValue(t) }

// ExtractBulkFor extracts the @bulk-for annotation value from a comment.
func ExtractBulkFor(comment string) string { return extractBulkFor(comment) }

// ToSingular converts a plural word to singular form.
func ToSingular(s string) string { return toSingular(s) }

// JoinParamsSignature joins parameters into a function signature string.
func JoinParamsSignature(params []Param) string { return joinParamsSignature(params) }

// JoinReturns joins return types into a comma-separated string.
func JoinReturns(returns []Return) string { return joinReturns(returns) }

// IsSlice checks if a type string represents a slice.
func IsSlice(retType string) bool { return isSlice(retType) }

// FirstReturnType returns the first return type from a Returns slice.
func FirstReturnType(returns []Return) string { return firstReturnType(returns) }

// HasParam checks if a parameter with the given name exists.
func HasParam(name string, params []Param) bool { return hasParam(name, params) }

// ParamHasField checks if a parameter's struct type has a given field.
func ParamHasField(paramName, fieldName string, params []Param, structs map[string]StructInfo) bool {
	return paramHasField(paramName, fieldName, params, structs)
}

// HasSliceField checks if a struct has a slice field.
func HasSliceField(s StructInfo) bool { return hasSliceField(s) }

// GetSliceField returns the first slice field of a struct.
func GetSliceField(s StructInfo) FieldInfo { return getSliceField(s) }

// ToSnakeCase converts a CamelCase string to snake_case.
func ToSnakeCase(s string) string { return toSnakeCase(s) }

// Quote wraps a string in engine-appropriate quotes.
func Quote(e Engine, s string) string { return quote(e, s) }

// GenerateFieldConversion generates field conversion code.
func GenerateFieldConversion(targetFieldName, targetFieldType, sourceFieldType, sourceExpr string) string {
	return generateFieldConversion(targetFieldName, targetFieldType, sourceFieldType, sourceExpr)
}

// WrapperTemplate is the template for generating wrapper files.
const WrapperTemplate = wrapperTemplate
