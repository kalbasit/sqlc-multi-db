package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
)

// Run is the main entry point for the generator.
// querierPath is the path to the source querier.go file (e.g., postgresdb/querier.go).
func Run(querierPath string) {
	engines := []Engine{
		{Name: "sqlite", Package: "sqlitedb"},
		{Name: "postgres", Package: "postgresdb"},
		{Name: "mysql", Package: "mysqldb"},
	}

	absQuerierPath, err := filepath.Abs(querierPath)
	if err != nil {
		log.Fatalf("resolving querier path: %v", err)
	}

	sourceDir := filepath.Dir(absQuerierPath)
	targetDir := filepath.Dir(sourceDir) // Parent of postgresdb is pkg/database

	// 1. Parse source package
	sourceData := parsePackage(sourceDir)

	// 2. Identify used structs from source methods
	usedStructNames := make(map[string]bool)

	for _, m := range sourceData.Methods {
		for _, p := range m.Params {
			cleanType := strings.TrimPrefix(p.Type, "[]")
			if _, exists := sourceData.Structs[cleanType]; exists {
				usedStructNames[cleanType] = true
			}
		}

		for _, r := range m.Returns {
			cleanType := strings.TrimPrefix(r.Type, "[]")
			if _, exists := sourceData.Structs[cleanType]; exists {
				usedStructNames[cleanType] = true
			}
		}
	}

	sortedStructs := make([]StructInfo, 0, len(usedStructNames))
	for name := range usedStructNames {
		sortedStructs = append(sortedStructs, sourceData.Structs[name])
	}

	sort.Slice(sortedStructs, func(i, j int) bool {
		return sortedStructs[i].Name < sortedStructs[j].Name
	})

	// 3. Synthesize missing GetByID methods
	for name := range sourceData.Structs {
		if !isDomainStruct(name) || strings.HasSuffix(name, "Params") || strings.HasSuffix(name, "Row") {
			continue
		}

		hasID := false

		for _, f := range sourceData.Structs[name].Fields {
			if f.Name == "ID" {
				hasID = true

				break
			}
		}

		if !hasID {
			continue
		}

		methodName := "Get" + name + "ByID"
		found := false

		for _, m := range sourceData.Methods {
			if m.Name == methodName {
				found = true

				break
			}
		}

		if !found {
			log.Printf("Synthesizing %s\n", methodName)
			sourceData.Methods = append(sourceData.Methods, MethodInfo{
				Name: methodName,
				Params: []Param{
					{Name: "ctx", Type: "context.Context"},
					{Name: "id", Type: "int64"},
				},
				Returns: []Return{
					{Type: name},
					{Type: "error"},
				},
				ReturnElem:   name,
				ReturnsError: true,
				HasValue:     true,
				IsSynthetic:  true,
				Docs:         []string{"// " + methodName + " (Synthetic)"},
			})
		}
	}

	sort.Slice(sourceData.Methods, func(i, j int) bool {
		return sourceData.Methods[i].Name < sourceData.Methods[j].Name
	})

	// 4. Detect package name and import base
	packageName := detectPackageName(targetDir)
	importBase := findImportBase(targetDir)

	// 5. Generate models.go, querier.go, and errors.go
	generateModels(targetDir, packageName, sortedStructs)
	generateQuerier(targetDir, packageName, sourceData.Methods)
	generateErrors(targetDir, packageName)

	// 6. Parse all target packages
	engineData := make(map[string]PackageData)

	for _, engine := range engines {
		engineDir := filepath.Join(targetDir, engine.Package)
		engineData[engine.Name] = parsePackage(engineDir)
	}

	// 7. Generate wrappers
	for _, engine := range engines {
		generateWrapper(
			targetDir, packageName, importBase, engine,
			sourceData.Methods, sourceData.Structs, engineData[engine.Name],
		)
	}
}

func parseQuerierInterface(typeSpec *ast.TypeSpec) ([]MethodInfo, bool) {
	if typeSpec.Name.Name != typeQuerier {
		return nil, false
	}

	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, false
	}

	methods := make([]MethodInfo, 0, len(interfaceType.Methods.List))

	for _, field := range interfaceType.Methods.List {
		m := MethodInfo{Name: field.Names[0].Name}
		if field.Doc != nil {
			for _, comment := range field.Doc.List {
				m.Docs = append(m.Docs, comment.Text)
				if strings.Contains(comment.Text, "@bulk-for") {
					if bulkFor := extractBulkFor(comment.Text); bulkFor != "" {
						m.BulkFor = bulkFor
					}
				}
			}
		}

		funcType := field.Type.(*ast.FuncType)
		for _, param := range funcType.Params.List {
			typeStr := exprToString(param.Type)
			for _, name := range param.Names {
				m.Params = append(m.Params, Param{Name: name.Name, Type: typeStr})
			}
		}

		if funcType.Results != nil {
			for _, res := range funcType.Results.List {
				typeStr := exprToString(res.Type)

				m.Returns = append(m.Returns, Return{Type: typeStr})
				switch typeStr {
				case "error":
					m.ReturnsError = true
				case typeQuerier:
					m.ReturnsSelf = true
					m.HasValue = true
				default:
					m.HasValue = true
					m.ReturnElem = strings.TrimPrefix(typeStr, "[]")
				}
			}
		}

		m.IsCreate = strings.HasPrefix(m.Name, "Create") && isDomainStruct(m.ReturnElem)
		m.IsUpdate = strings.HasPrefix(m.Name, "Update") && isDomainStruct(m.ReturnElem)
		methods = append(methods, m)
	}

	return methods, true
}

func parseStructType(typeSpec *ast.TypeSpec, structType *ast.StructType) StructInfo {
	s := StructInfo{Name: typeSpec.Name.Name}

	if structType.Fields == nil {
		return s
	}

	for _, field := range structType.Fields.List {
		typeStr := exprToString(field.Type)
		tag := ""

		if field.Tag != nil {
			unquoted, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				log.Fatalf("failed to unquote struct tag %s: %v", field.Tag.Value, err)
			}

			tag = unquoted
		}

		if len(field.Names) > 0 {
			for _, name := range field.Names {
				s.Fields = append(s.Fields, FieldInfo{Name: name.Name, Type: typeStr, Tag: tag})
			}
		} else {
			s.Fields = append(s.Fields, FieldInfo{Name: "", Type: typeStr, Tag: tag})
		}
	}

	return s
}

func parsePackage(dir string) PackageData {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	methods := make([]MethodInfo, 0, 32)

	structs := make(map[string]StructInfo)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				if querierMethods, matched := parseQuerierInterface(typeSpec); matched {
					methods = append(methods, querierMethods...)
				}

				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					s := parseStructType(typeSpec, structType)
					structs[s.Name] = s
				}

				return true
			})
		}
	}

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	return PackageData{Methods: methods, Structs: structs}
}

func generateModels(dir, packageName string, structs []StructInfo) {
	t := template.Must(template.New("models").Parse(modelsTemplate))

	var buf bytes.Buffer

	data := map[string]interface{}{
		"PackageName": packageName,
		"Structs":     structs,
	}
	if err := t.Execute(&buf, data); err != nil {
		log.Fatalf("executing models template: %v", err)
	}

	writeFile(dir, generatedFilePrefix+"models.go", buf.Bytes())
}

func generateQuerier(dir, packageName string, methods []MethodInfo) {
	t := template.Must(template.New("querier").Funcs(template.FuncMap{
		"joinParamsSignature": joinParamsSignature,
		"joinReturns":         joinReturns,
	}).Parse(querierTemplate))

	var buf bytes.Buffer

	data := map[string]interface{}{
		"PackageName": packageName,
		"Methods":     methods,
	}
	if err := t.Execute(&buf, data); err != nil {
		log.Fatalf("executing querier template: %v", err)
	}

	writeFile(dir, generatedFilePrefix+"querier.go", buf.Bytes())
}

func generateErrors(dir, packageName string) {
	t := template.Must(template.New("errors").Parse(errorsTemplate))

	var buf bytes.Buffer

	data := map[string]interface{}{
		"PackageName": packageName,
	}
	if err := t.Execute(&buf, data); err != nil {
		log.Fatalf("executing errors template: %v", err)
	}

	writeFile(dir, generatedFilePrefix+"errors.go", buf.Bytes())
}

func generateWrapper(
	dir, packageName, importBase string,
	engine Engine,
	methods []MethodInfo,
	structs map[string]StructInfo,
	engData PackageData,
) {
	t := template.Must(template.New("wrapper").Funcs(template.FuncMap{
		"joinParamsSignature": joinParamsSignature,
		"joinReturns":         joinReturns,
		"isSlice":             isSlice,
		"firstReturnType":     firstReturnType,
		"isDomainStruct":      isDomainStructFunc,
		"zeroValue":           zeroValue,
		"getStruct":           func(name string) StructInfo { return structs[name] },
		"hasSliceField":       hasSliceField,
		"getSliceField":       getSliceField,
		"toSingular":          toSingular,
		"trimPrefix":          strings.TrimPrefix,
		"getTargetMethod": func(name string) MethodInfo {
			for _, m := range engData.Methods {
				if m.Name == name {
					return m
				}
			}

			return MethodInfo{}
		},
		"getTargetStruct": func(name string) StructInfo {
			if engData.Structs == nil {
				return StructInfo{}
			}

			return engData.Structs[name]
		},
		"joinParamsCall": func(params []Param, engPkg string, targetMethodName string) (string, error) {
			targetMethod := MethodInfo{}

			if engData.Methods != nil {
				for _, m := range engData.Methods {
					if m.Name == targetMethodName {
						targetMethod = m

						break
					}
				}
			}

			return joinParamsCall(params, engPkg, targetMethod, engData.Structs, structs)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errInvalidDictCall
			}

			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errDictKeysMustBeStrings
				}

				dict[key] = values[i+1]
			}

			return dict, nil
		},
		"hasSuffix":               strings.HasSuffix,
		"toSnakeCase":             toSnakeCase,
		"quote":                   quote,
		"generateFieldConversion": generateFieldConversion,
		"hasParam":                hasParam,
		"paramHasField":           paramHasField,
		"getTableName": func(structName string) string {
			extractTableName := func(docs []string) (string, bool) {
				clauses := []struct {
					keyword string
					offset  int
				}{
					{"INSERT INTO ", 12},
					{"UPDATE ", 7},
					{"DELETE FROM ", 12},
					{"FROM ", 5},
				}

				for _, doc := range docs {
					doc = strings.ToUpper(doc)
					for _, clause := range clauses {
						if idx := strings.Index(doc, clause.keyword); idx != -1 {
							parts := strings.Fields(doc[idx+clause.offset:])
							if len(parts) > 0 {
								tableName := strings.ToLower(parts[0])
								if clause.keyword == "INSERT INTO " {
									tableName = strings.Trim(tableName, "()")
								}

								return tableName, true
							}
						}
					}
				}

				return "", false
			}

			prefixes := []string{"Create", "Update", "Get", "Delete"}
			for _, p := range prefixes {
				for _, m := range methods {
					if m.Name == p+structName || strings.HasPrefix(m.Name, p+structName) {
						if tableName, found := extractTableName(m.Docs); found {
							return tableName
						}
					}
				}
			}

			for _, m := range methods {
				if strings.Contains(m.Name, structName) {
					if tableName, found := extractTableName(m.Docs); found {
						return tableName
					}
				}
			}

			return strings.ToLower(inflection.Plural(structName))
		},
	}).Parse(wrapperTemplate))

	var buf bytes.Buffer

	data := map[string]interface{}{
		"Engine":      engine,
		"Methods":     methods,
		"Structs":     structs,
		"ImportBase":  importBase,
		"PackageName": packageName,
	}

	if err := t.Execute(&buf, data); err != nil {
		log.Fatalf("executing wrapper template: %v", err)
	}

	writeFile(dir, fmt.Sprintf("%swrapper_%s.go", generatedFilePrefix, engine.Name), buf.Bytes())
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.InterfaceType:
		return typeAny
	default:
		panic(fmt.Sprintf("unhandled expression type: %T", t))
	}
}
