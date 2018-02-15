package generator

type Field struct {
	Name
	DataType          string
	DataTypeIn        string
	DataTypeDB        string
	IsRepeatedBuiltin bool
	IsStruct          bool
	IsRepeatedStruct  bool
	IsMap             bool
	Tag               string
	// DataTypeInName Name
	// MapKeyDataType   string
	// MapValueDataType string
	// Transform        string
	// Validate         string
	// Rule             string
}
