package generator

type Genni struct {
	Backtick string
	Package  string
	Structs  []Struct
}

func NewGenni() (genni Genni) {
	genni = Genni{
		Backtick: "`",
		Structs:  []Struct{},
	}
	return
}
