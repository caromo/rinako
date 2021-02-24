package collections

//User represents information about a user pertinent to Rinako
type User struct {
	ID       string `gorm: primary_key`
	Names    []Name
	Targeted bool
}

//Name represents a name
type Name struct {
	Name string
}
