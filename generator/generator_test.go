package generator_test

import (
	"bytes"
	"errors"
	"go/ast"
	"strings"
	"testing"
	"text/template"

	"github.com/kalbasit/sqlc-multi-db/generator"
)

var (
	errTestInvalidDictCall       = errors.New("invalid dict call")
	errTestDictKeysMustBeStrings = errors.New("dict keys must be strings")
)

func TestExprToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expr     ast.Expr
		expected string
		panics   bool
	}{
		{
			name:     "Ident",
			expr:     &ast.Ident{Name: "int"},
			expected: "int",
		},
		{
			name:     "StarExpr",
			expr:     &ast.StarExpr{X: &ast.Ident{Name: "String"}},
			expected: "*String",
		},
		{
			name:     "ArrayType",
			expr:     &ast.ArrayType{Elt: &ast.Ident{Name: "byte"}},
			expected: "[]byte",
		},
		{
			name:     "SelectorExpr",
			expr:     &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "NullString"}},
			expected: "sql.NullString",
		},
		{
			name:   "Unhandled MapType",
			expr:   &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}},
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if r := recover(); r != nil {
					if !tt.panics {
						t.Errorf("ExprToString panicked unexpectedly: %v", r)
					}
				} else if tt.panics {
					t.Errorf("ExprToString expected panic but did not panic")
				}
			}()

			result := generator.ExprToString(tt.expr)
			if !tt.panics && result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsDomainStructFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputType string
		want      bool
	}{
		{"User", true},
		{"[]User", true},
		{"sql.NullString", false},
		{"int", false},
		{"string", false},
		{"Querier", false},
		{"pkg.User", false},
		{"user", false},
	}

	for _, tt := range tests {
		if got := generator.IsDomainStructFunc(tt.inputType); got != tt.want {
			t.Errorf("IsDomainStructFunc(%q) = %v, want %v", tt.inputType, got, tt.want)
		}
	}
}

func TestZeroValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		typeName string
		want     string
	}{
		{"int", "0"},
		{"string", `""`},
		{"bool", "false"},
		{"error", "nil"},
		{"*User", "nil"},
		{"[]byte", "nil"},
		{"MyStruct", "MyStruct{}"},
	}

	for _, tt := range tests {
		if got := generator.ZeroValue(tt.typeName); got != tt.want {
			t.Errorf("ZeroValue(%q) = %q, want %q", tt.typeName, got, tt.want)
		}
	}
}

func TestExtractBulkFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		comment string
		want    string
	}{
		{"// CreateUsers creates users @bulk-for CreateUser", "CreateUser"},
		{"// @bulk-for CreateUser", "CreateUser"},
		{"// No annotation here", ""},
		{"// Multiple @bulk-for First @bulk-for Second", "First"},
		{"// @bulk-for", ""},
	}

	for _, tt := range tests {
		if got := generator.ExtractBulkFor(tt.comment); got != tt.want {
			t.Errorf("ExtractBulkFor(%q) = %q, want %q", tt.comment, got, tt.want)
		}
	}
}

func TestToSingular(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"Users", "User"},
		{"Process", "Process"},
		{"GetStatus", "GetStatus"},
		{"Status", "Status"},
		{"Addresses", "Address"},
	}

	for _, tt := range tests {
		if got := generator.ToSingular(tt.input); got != tt.want {
			t.Errorf("ToSingular(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestJoinParamsCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  []generator.Param
		engPkg  string
		want    string
		wantErr bool
	}{
		{
			name: "Simple Params",
			params: []generator.Param{
				{Name: "ctx", Type: "context.Context"},
				{Name: "id", Type: "int64"},
			},
			engPkg: "sqlitedb",
			want:   "ctx, id",
		},
		{
			name: "Domain Struct Param",
			params: []generator.Param{
				{Name: "user", Type: "User"},
			},
			engPkg: "postgresdb",
			want:   "postgresdb.User(user)",
		},
		{
			name: "Unsupported Slice of Domain Struct",
			params: []generator.Param{
				{Name: "users", Type: "[]User"},
			},
			engPkg:  "postgresdb",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := generator.JoinParamsCall(tt.params, tt.engPkg, generator.MethodInfo{}, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinParamsCall() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("JoinParamsCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapperTemplate(t *testing.T) {
	t.Parallel()

	// Mock engines
	sqlite := generator.Engine{Name: "sqlite", Package: "sqlitedb"}

	// Mock structs
	structs := map[string]generator.StructInfo{
		"CreateUserParams": {
			Name: "CreateUserParams",
			Fields: []generator.FieldInfo{
				{Name: "Username", Type: "string"},
			},
		},
		"CreateUsersParams": {
			Name: "CreateUsersParams",
			Fields: []generator.FieldInfo{
				{Name: "Usernames", Type: "[]string"},
			},
		},
	}

	// Mock methods
	methods := []generator.MethodInfo{
		{
			Name: "CreateUsers",
			Params: []generator.Param{
				{Name: "ctx", Type: "context.Context"},
				{Name: "arg", Type: "CreateUsersParams"},
			},
			Returns: []generator.Return{{Type: "error"}},
			Docs:    []string{"// CreateUsers creates users"},
		},
	}

	funcMap := template.FuncMap{
		"hasParam":            generator.HasParam,
		"paramHasField":       generator.ParamHasField,
		"joinParamsSignature": generator.JoinParamsSignature,
		"joinReturns":         generator.JoinReturns,
		"isSlice":             generator.IsSlice,
		"firstReturnType":     generator.FirstReturnType,
		"isDomainStruct":      generator.IsDomainStructFunc,
		"zeroValue":           generator.ZeroValue,
		"getStruct":           func(name string) generator.StructInfo { return structs[name] },
		"hasSliceField":       generator.HasSliceField,
		"getSliceField":       generator.GetSliceField,
		"toSingular":          generator.ToSingular,
		"trimPrefix":          strings.TrimPrefix,
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errTestInvalidDictCall
			}

			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errTestDictKeysMustBeStrings
				}

				dict[key] = values[i+1]
			}

			return dict, nil
		},
		"getTargetMethod": func(name string) generator.MethodInfo {
			if name == "CreateUsers" {
				return generator.MethodInfo{
					Name: "CreateUsers",
					Params: []generator.Param{
						{Name: "ctx", Type: "context.Context"},
						{Name: "arg", Type: "CreateUsersParams"},
					},
					Returns: []generator.Return{{Type: "error"}},
				}
			}

			return generator.MethodInfo{}
		},
		"getTargetStruct": func(name string) generator.StructInfo { return structs[name] },
		"joinParamsCall": func(params []generator.Param, engPkg string, targetMethodName string) (string, error) {
			targetMethod := generator.MethodInfo{}
			if targetMethodName == "CreateUsers" {
				targetMethod = generator.MethodInfo{
					Name: "CreateUsers",
					Params: []generator.Param{
						{Name: "ctx", Type: "context.Context"},
						{Name: "arg", Type: "CreateUsersParams"},
					},
					Returns: []generator.Return{{Type: "error"}},
				}
			}

			return generator.JoinParamsCall(params, engPkg, targetMethod, structs, structs)
		},
		"hasSuffix":               strings.HasSuffix,
		"toSnakeCase":             generator.ToSnakeCase,
		"quote":                   generator.Quote,
		"generateFieldConversion": generator.GenerateFieldConversion,
		"zeroReturn": func(m generator.MethodInfo) string {
			if m.ReturnsSelf {
				return "nil"
			}

			return "0"
		},
		"getTableName": func(_ string) string { return "users" },
	}

	tmpl, err := template.New("wrapper").Funcs(funcMap).Parse(generator.WrapperTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"Engine":      sqlite,
		"Methods":     methods,
		"Structs":     structs,
		"ImportBase":  "github.com/example/project/pkg/database",
		"PackageName": "database",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()

	// Verify auto-looping was triggered
	if !strings.Contains(output, "for i, v := range arg.Usernames") {
		t.Errorf("expected output to contain loop over arg.Usernames, but it didn't\n%s", output)
	}

	// Verify field mapping by type
	if !strings.Contains(output, "Username: v,") {
		t.Errorf("expected output to contain 'Username: v,', but it didn't\n%s", output)
	}

	// 2. Test GetStatus (should NOT loop because GetStatuParams does not exist)
	methods = []generator.MethodInfo{
		{
			Name: "GetStatus",
			Params: []generator.Param{
				{Name: "ctx", Type: "context.Context"},
				{Name: "hash", Type: "string"},
			},
			Returns: []generator.Return{{Type: "Status"}, {Type: "error"}},
			Docs:    []string{"// GetStatus gets status"},
		},
	}

	data["Methods"] = methods

	buf.Reset()

	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output = buf.String()
	if strings.Contains(output, "for _, v := range") {
		t.Errorf("expected output NOT to contain loop for GetStatus, but it did\n%s", output)
	}

	// 3. Test ReturnsSelf (WithTx)
	methods = []generator.MethodInfo{
		{
			Name:         "WithTx",
			Params:       []generator.Param{{Name: "tx", Type: "*sql.Tx"}},
			Returns:      []generator.Return{{Type: "Querier"}, {Type: "error"}},
			ReturnsSelf:  true,
			ReturnsError: true,
			HasValue:     true,
			Docs:         []string{"// WithTx returns a new Querier with transaction"},
		},
	}

	data["Methods"] = methods

	buf.Reset()

	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "nil, ErrNotFound") {
		t.Errorf("expected output to contain 'nil, ErrNotFound' for WithTx, but it didn't\n%s", output)
	}

	if !strings.Contains(output, "nil, err") {
		t.Errorf("expected output to contain 'nil, err' for WithTx, but it didn't\n%s", output)
	}

	// 4. Test sql.NullString conversion
	methods = []generator.MethodInfo{
		{
			Name: "CreateUser",
			Params: []generator.Param{
				{Name: "ctx", Type: "context.Context"},
				{Name: "user", Type: "User"},
			},
			Returns: []generator.Return{{Type: "error"}},
		},
	}

	structs["CreateUserParams"] = generator.StructInfo{
		Name: "CreateUserParams",
		Fields: []generator.FieldInfo{
			{Name: "Bio", Type: "sql.NullString"},
		},
	}
	structs["User"] = generator.StructInfo{
		Name: "User",
		Fields: []generator.FieldInfo{
			{Name: "Bio", Type: "string"},
		},
	}

	funcMap["getTargetMethod"] = func(name string) generator.MethodInfo {
		if name == "CreateUser" {
			return generator.MethodInfo{
				Name: "CreateUser",
				Params: []generator.Param{
					{Name: "ctx", Type: "context.Context"},
					{Name: "arg", Type: "CreateUserParams"},
				},
				Returns: []generator.Return{{Type: "error"}},
			}
		}

		return generator.MethodInfo{}
	}

	funcMap["joinParamsCall"] = func(params []generator.Param, engPkg string, _ string) (string, error) {
		targetMethod := generator.MethodInfo{
			Name: "CreateUser",
			Params: []generator.Param{
				{Name: "ctx", Type: "context.Context"},
				{Name: "arg", Type: "CreateUserParams"},
			},
			Returns: []generator.Return{{Type: "error"}},
		}

		return generator.JoinParamsCall(params, engPkg, targetMethod, structs, structs)
	}
	funcMap["getTableName"] = func(_ string) string { return "users" }

	tmpl, err = template.New("wrapper").Funcs(funcMap).Parse(generator.WrapperTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data["Methods"] = methods

	buf.Reset()

	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output = buf.String()

	expectedConversion := "Bio: sql.NullString{String: user.Bio, Valid: true}"
	if !strings.Contains(output, expectedConversion) {
		t.Errorf("expected output to contain '%s', but it didn't\n%s", expectedConversion, output)
	}
}

func TestGenerateFieldConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		targetFieldName string
		targetFieldType string
		sourceFieldType string
		sourceExpr      string
		want            string
	}{
		{
			name:            "Same Types",
			targetFieldName: "ID",
			targetFieldType: "int64",
			sourceFieldType: "int64",
			sourceExpr:      "user.ID",
			want:            "ID: user.ID",
		},
		{
			name:            "String to NullString",
			targetFieldName: "Bio",
			targetFieldType: "sql.NullString",
			sourceFieldType: "string",
			sourceExpr:      "user.Bio",
			want:            "Bio: sql.NullString{String: user.Bio, Valid: true}",
		},
		{
			name:            "Int64 to NullInt64",
			targetFieldName: "Age",
			targetFieldType: "sql.NullInt64",
			sourceFieldType: "int64",
			sourceExpr:      "user.Age",
			want:            "Age: sql.NullInt64{Int64: user.Age, Valid: true}",
		},
		{
			name:            "NullString to String",
			targetFieldName: "Bio",
			targetFieldType: "string",
			sourceFieldType: "sql.NullString",
			sourceExpr:      "row.Bio",
			want:            "Bio: row.Bio.String",
		},
		{
			name:            "NullInt32 to NullInt64",
			targetFieldName: "Count",
			targetFieldType: "sql.NullInt64",
			sourceFieldType: "sql.NullInt32",
			sourceExpr:      "src.Count",
			want:            "Count: sql.NullInt64{Int64: int64(src.Count.Int32), Valid: src.Count.Valid}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := generator.GenerateFieldConversion(tt.targetFieldName, tt.targetFieldType, tt.sourceFieldType, tt.sourceExpr)
			if got != tt.want {
				t.Errorf("GenerateFieldConversion() = %v, want %v", got, tt.want)
			}
		})
	}
}
