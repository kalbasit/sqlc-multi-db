package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/inflection"
	"golang.org/x/tools/imports"
	"mvdan.cc/gofumpt/format"
)

func toSnakeCase(s string) string {
	res := make([]rune, 0, len(s))

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if previous was also uppercase (e.g. ID)
			prev := rune(s[i-1])
			if prev < 'A' || prev > 'Z' {
				res = append(res, '_')
			}
		}

		res = append(res, []rune(strings.ToLower(string(r)))[0])
	}

	return string(res)
}

func quote(e Engine, s string) string {
	if e.IsMySQL() {
		return "`" + s + "`"
	}

	return `\"` + s + `\"`
}

func extractBulkFor(comment string) string {
	parts := strings.Fields(comment)
	for i, p := range parts {
		if p == "@bulk-for" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

func toSingular(s string) string { return inflection.Singular(s) }

func writeFile(dir, filename string, content []byte) {
	// 1. Manage imports with goimports
	withImports, err := imports.Process(filename, content, nil)
	if err != nil {
		log.Println(string(content))
		log.Fatalf("imports.Process %s: %v", filename, err)
	}

	// 2. Format with gofumpt
	formatted, err := format.Source(withImports, format.Options{
		LangVersion: "",
		ExtraRules:  true,
	})
	if err != nil {
		log.Println(string(withImports))
		log.Fatalf("formatting %s: %v", filename, err)
	}

	if err := os.WriteFile(filepath.Join(dir, filename), formatted, 0o644); err != nil { //nolint:gosec
		log.Fatal(err)
	}

	fmt.Printf("Generated %s\n", filename)
}

// hasParam checks if a parameter with the given name exists in the params list.
func hasParam(name string, params []Param) bool {
	for _, param := range params {
		if param.Name == name {
			return true
		}
	}

	return false
}

func paramHasField(paramName string, fieldName string, params []Param, structs map[string]StructInfo) bool {
	for _, param := range params {
		if param.Name == paramName {
			typeName := strings.TrimPrefix(param.Type, "[]")
			typeName = strings.TrimPrefix(typeName, "*")

			typeParts := strings.Split(typeName, ".")
			if len(typeParts) > 1 {
				typeName = typeParts[len(typeParts)-1]
			}

			if s, ok := structs[typeName]; ok {
				for _, f := range s.Fields {
					if f.Name == fieldName {
						return true
					}
				}
			}

			return false
		}
	}

	return false
}

func joinParamsSignature(params []Param) string {
	p := make([]string, 0, len(params))
	for _, param := range params {
		p = append(p, fmt.Sprintf("%s %s", param.Name, param.Type))
	}

	return strings.Join(p, ", ")
}

// JoinParamsCall is exported for use in tests.
func JoinParamsCall(
	params []Param,
	engPkg string,
	targetMethod MethodInfo,
	targetStructs map[string]StructInfo,
	sourceStructs map[string]StructInfo,
) (string, error) {
	return joinParamsCall(params, engPkg, targetMethod, targetStructs, sourceStructs)
}

func joinDomainStructParam(
	param Param,
	i int,
	engPkg string,
	targetMethod MethodInfo,
	targetStructs map[string]StructInfo,
	sourceStructs map[string]StructInfo,
) (string, error) {
	if strings.HasPrefix(param.Type, "[]") {
		return "", errUnsupportedSliceDomainStruct(param.Type)
	}

	targetParamType := ""
	if i < len(targetMethod.Params) {
		targetParamType = targetMethod.Params[i].Type
	}

	if targetParamType != "" {
		sourceStruct := sourceStructs[param.Type]
		targetStruct := targetStructs[targetParamType]

		var fields []string

		for _, targetField := range targetStruct.Fields {
			var sourceField FieldInfo

			found := false

			for _, sf := range sourceStruct.Fields {
				if sf.Name == targetField.Name {
					sourceField = sf
					found = true

					break
				}
			}

			if found {
				conversion := generateFieldConversion(
					targetField.Name,
					targetField.Type,
					sourceField.Type,
					fmt.Sprintf("%s.%s", param.Name, sourceField.Name),
				)
				fields = append(fields, conversion)
			}
		}

		return fmt.Sprintf("%s.%s{\n%s,\n}", engPkg, targetParamType, strings.Join(fields, ",\n")), nil
	}

	return fmt.Sprintf("%s.%s(%s)", engPkg, param.Type, param.Name), nil
}

func joinNonDomainParam(param Param, i int, targetMethod MethodInfo) string {
	targetParamType := ""
	if i < len(targetMethod.Params) {
		targetParamType = targetMethod.Params[i].Type
	}

	if targetParamType != "" && targetParamType != param.Type {
		return fmt.Sprintf("%s(%s)", targetParamType, param.Name)
	}

	return param.Name
}

func joinParamsCall(
	params []Param,
	engPkg string,
	targetMethod MethodInfo,
	targetStructs map[string]StructInfo,
	sourceStructs map[string]StructInfo,
) (string, error) {
	p := make([]string, 0, len(params))

	for i, param := range params {
		if isDomainStructFunc(param.Type) {
			result, err := joinDomainStructParam(param, i, engPkg, targetMethod, targetStructs, sourceStructs)
			if err != nil {
				return "", err
			}

			p = append(p, result)
		} else {
			p = append(p, joinNonDomainParam(param, i, targetMethod))
		}
	}

	return strings.Join(p, ", "), nil
}

func joinReturns(returns []Return) string {
	r := make([]string, 0, len(returns))
	for _, ret := range returns {
		r = append(r, ret.Type)
	}

	return strings.Join(r, ", ")
}

func isSlice(retType string) bool {
	return strings.HasPrefix(retType, "[]")
}

func firstReturnType(returns []Return) string {
	if len(returns) > 0 {
		return returns[0].Type
	}

	return ""
}

// isDomainStructFunc checks if type is a "Domain Struct" based on naming convention.
func isDomainStructFunc(t string) bool {
	t = strings.TrimPrefix(t, "[]")

	return len(t) > 0 && t[0] >= 'A' && t[0] <= 'Z' && !strings.Contains(t, ".") && t != typeQuerier
}

// isDomainStruct is used during parsing, same logic.
func isDomainStruct(t string) bool {
	return isDomainStructFunc(t)
}

func zeroValue(t string) string {
	if isNumeric(t) {
		return "0"
	}

	switch t {
	case typeBool:
		return "false"
	case typeString:
		return `""`
	case "error":
		return zeroNil
	}

	if strings.HasPrefix(t, "*") || strings.HasPrefix(t, "[]") || strings.HasPrefix(t, "map[") || t == typeAny {
		return zeroNil
	}

	if t == "sql.Result" || t == typeQuerier {
		return zeroNil
	}

	return fmt.Sprintf("%s{}", t)
}

func isNumeric(t string) bool {
	switch t {
	case "int", "int8", typeInt16, typeInt32, typeInt64:
		return true
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "float32", typeFloat64, "complex64", "complex128":
		return true
	case typeByte, "rune":
		return true
	}

	return false
}

func isStructType(t string) bool {
	return strings.HasPrefix(t, "sql.Null")
}

func isSQLNullType(t string) bool {
	return strings.HasPrefix(t, "sql.Null")
}

func getPrimitiveFromNullType(t string) string {
	switch t {
	case sqlNullString:
		return typeString
	case sqlNullInt64:
		return typeInt64
	case sqlNullInt32:
		return typeInt32
	case sqlNullInt16:
		return typeInt16
	case sqlNullBool:
		return typeBool
	case sqlNullFloat64:
		return typeFloat64
	case sqlNullTime:
		return "time.Time"
	case sqlNullByte:
		return typeByte
	default:
		return ""
	}
}

func getNullTypeFromPrimitive(t string) string {
	switch t {
	case typeString:
		return sqlNullString
	case typeInt64:
		return sqlNullInt64
	case typeInt32:
		return sqlNullInt32
	case typeInt16:
		return sqlNullInt16
	case typeBool:
		return sqlNullBool
	case typeFloat64:
		return sqlNullFloat64
	case "time.Time":
		return sqlNullTime
	case typeByte:
		return sqlNullByte
	default:
		return ""
	}
}

func getFieldNameForNullType(t string) string {
	switch t {
	case sqlNullString:
		return "String"
	case sqlNullInt64:
		return "Int64"
	case sqlNullInt32:
		return "Int32"
	case sqlNullInt16:
		return "Int16"
	case sqlNullBool:
		return "Bool"
	case sqlNullFloat64:
		return "Float64"
	case sqlNullTime:
		return "Time"
	case sqlNullByte:
		return "Byte"
	default:
		return ""
	}
}

// generateFieldConversion generates the conversion code for a field mapping.
func generateFieldConversion(targetFieldName, targetFieldType, sourceFieldType, sourceExpr string) string {
	// Case 1: Types are identical - direct assignment
	if sourceFieldType == targetFieldType {
		return fmt.Sprintf("%s: %s", targetFieldName, sourceExpr)
	}

	// Case 4: Both are sql.Null* types but different
	if isSQLNullType(sourceFieldType) && isSQLNullType(targetFieldType) {
		sourcePrimitive := getPrimitiveFromNullType(sourceFieldType)

		targetPrimitive := getPrimitiveFromNullType(targetFieldType)
		if sourcePrimitive != "" && targetPrimitive != "" {
			sourceFieldName := getFieldNameForNullType(sourceFieldType)

			targetValueFieldName := getFieldNameForNullType(targetFieldType)
			if sourcePrimitive == targetPrimitive {
				return fmt.Sprintf(
					"%s: %s{%s: %s.%s, Valid: %s.Valid}",
					targetFieldName, targetFieldType, targetValueFieldName,
					sourceExpr, sourceFieldName, sourceExpr,
				)
			}

			return fmt.Sprintf(
				"%s: %s{%s: %s(%s.%s), Valid: %s.Valid}",
				targetFieldName, targetFieldType, targetValueFieldName,
				targetPrimitive, sourceExpr, sourceFieldName, sourceExpr,
			)
		}
	}

	// Case 2: Converting from primitive to sql.Null* (skip interface{} — handled by Case 5b)
	if isSQLNullType(targetFieldType) && sourceFieldType != typeAny {
		expectedPrimitive := getPrimitiveFromNullType(targetFieldType)
		if expectedPrimitive == sourceFieldType {
			fieldName := getFieldNameForNullType(targetFieldType)

			return fmt.Sprintf("%s: %s{%s: %s, Valid: true}", targetFieldName, targetFieldType, fieldName, sourceExpr)
		} else if expectedPrimitive != "" {
			fieldName := getFieldNameForNullType(targetFieldType)

			return fmt.Sprintf(
				"%s: %s{%s: %s(%s), Valid: true}",
				targetFieldName, targetFieldType, fieldName, expectedPrimitive, sourceExpr,
			)
		}
	}

	// Case 3: Converting from sql.Null* to primitive
	if isSQLNullType(sourceFieldType) {
		primitive := getPrimitiveFromNullType(sourceFieldType)
		if primitive == targetFieldType {
			fieldName := getFieldNameForNullType(sourceFieldType)

			return fmt.Sprintf("%s: %s.%s", targetFieldName, sourceExpr, fieldName)
		} else if primitive != "" {
			fieldName := getFieldNameForNullType(sourceFieldType)

			return fmt.Sprintf("%s: %s(%s.%s)", targetFieldName, targetFieldType, sourceExpr, fieldName)
		}
	}

	// Case 5: Struct types (non-sql.Null*) - direct assignment
	if isStructType(targetFieldType) {
		return fmt.Sprintf("%s: %s", targetFieldName, sourceExpr)
	}

	// Case 5b: interface{} source → sql.Null* target (SQLite nullable columns come as interface{})
	if sourceFieldType == typeAny && isSQLNullType(targetFieldType) {
		primitive := getPrimitiveFromNullType(targetFieldType)

		fieldName := getFieldNameForNullType(targetFieldType)
		if primitive != "" && fieldName != "" {
			return fmt.Sprintf(
				"%s: func() %s { if %s == nil { return %s{} }; v, ok := %s.(%s); if !ok { return %s{} };"+
					" return %s{%s: v, Valid: true} }()",
				targetFieldName, targetFieldType,
				sourceExpr, targetFieldType,
				sourceExpr, primitive,
				targetFieldType, targetFieldType, fieldName,
			)
		}
	}

	// Case 6: Primitive type conversion
	return fmt.Sprintf("%s: %s(%s)", targetFieldName, targetFieldType, sourceExpr)
}

func hasSliceField(s StructInfo) bool {
	for _, f := range s.Fields {
		if strings.HasPrefix(f.Type, "[]") && f.Type != "[]byte" {
			return true
		}
	}

	return false
}

func getSliceField(s StructInfo) FieldInfo {
	for _, f := range s.Fields {
		if strings.HasPrefix(f.Type, "[]") && f.Type != "[]byte" {
			return f
		}
	}

	return FieldInfo{}
}

func parseGoMod(goModPath, targetDir string) string {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		log.Fatalf("reading go.mod at %s: %v", goModPath, err)
	}

	moduleName := ""

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))

			break
		}
	}

	if moduleName == "" {
		log.Fatalf("could not find module directive in %s", goModPath)
	}

	dir := filepath.Dir(goModPath)

	relPath, err := filepath.Rel(dir, targetDir)
	if err != nil {
		log.Fatalf("computing relative path: %v", err)
	}

	if relPath == "." {
		return moduleName
	}

	return moduleName + "/" + relPath
}

// findImportBase walks up from targetDir to find the nearest go.mod and computes
// the full import path for targetDir.
func findImportBase(targetDir string) string {
	dir := targetDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return parseGoMod(goModPath, targetDir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatalf("no go.mod found walking up from %s", targetDir)
		}

		dir = parent
	}
}

// detectPackageName scans .go files in dir (skipping generated_*.go) to find the package name.
func detectPackageName(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return filepath.Base(dir)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		if strings.HasPrefix(name, "generated_") {
			continue
		}

		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package ") {
				pkg := strings.TrimSpace(strings.TrimPrefix(line, "package "))
				// Remove any trailing comment
				if idx := strings.Index(pkg, " "); idx != -1 {
					pkg = pkg[:idx]
				}

				return pkg
			}
		}
	}

	return filepath.Base(dir)
}

// Ensure getNullTypeFromPrimitive is used (referenced in templates indirectly).
var _ = getNullTypeFromPrimitive
