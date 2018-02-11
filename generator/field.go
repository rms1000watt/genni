package generator

type Field struct {
	Name
	DataTypeName Name
	DataType     string
	DataTypeDB   string
	// MapKeyDataType   string
	// MapValueDataType string
	// Transform        string
	// Validate         string
	// Rule             string
	IsRepeated       bool
	IsStruct         bool
	IsRepeatedStruct bool

	Tag string
}
