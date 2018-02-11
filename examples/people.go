package examples

type PersonIn struct {
	ID          int `json:"id" db:"id"`
	FirstName   string
	LastName    string
	BrotherMSPo map[string]PersonOut
	BrotherMSS  map[string]string
	BrotherAPo  []PersonOut
	BrotherPo   PersonOut
}

type PersonOut struct {
	FirstName string
	LastName  string
}
