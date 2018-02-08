package examples

type PersonIn struct {
	ID        int `json:"id"`
	FirstName string
	LastName  string
}

type PersonOut struct {
	FirstName string
	LastName  string
}
