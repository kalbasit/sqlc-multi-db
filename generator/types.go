package generator

// Engine configuration
type Engine struct {
	Name    string // e.g. "sqlite"
	Package string // e.g. "sqlitedb"
}

func (e Engine) IsMySQL() bool    { return e.Name == "mysql" }
func (e Engine) IsPostgres() bool { return e.Name == "postgres" }

// MethodInfo holds extracted data from the AST
type MethodInfo struct {
	Name         string
	Params       []Param
	Returns      []Return
	IsCreate     bool   // Special handling for MySQL Create
	IsUpdate     bool   // Special handling for MySQL Update
	ReturnElem   string // The underlying type (e.g. "NarFile" or "string")
	ReturnsError bool   // Does the method return an error?
	ReturnsSelf  bool   // Does it return the wrapper type (like WithTx)?
	HasValue     bool   // Does it return a value (non-error)?
	Docs         []string
	BulkFor      string // Extracted from @bulk-for annotation
	IsSynthetic  bool   // Is this method automatically generated?
}

type Param struct {
	Name string
	Type string
}

type Return struct {
	Type string
}

type StructInfo struct {
	Name   string
	Fields []FieldInfo
}

type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

type PackageData struct {
	Methods []MethodInfo
	Structs map[string]StructInfo
}
