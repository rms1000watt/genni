package generator

type Field struct {
	Name
	DataTypeIn       string
	DataTypeDB       string
	IsRepeated       bool
	IsStruct         bool
	IsRepeatedStruct bool
	IsMap            bool
	Tag              string
	// DataTypeInName Name
	// MapKeyDataType   string
	// MapValueDataType string
	// Transform        string
	// Validate         string
	// Rule             string
}
