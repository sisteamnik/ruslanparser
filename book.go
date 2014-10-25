package ruslanparser

type Book struct {
	SourceId string
	Title    string
	Author   string
	Series   string
	Tags     []string
	Genre    string

	Places []Place

	PageNum int

	City            string
	Publishing      string
	PublicationYear int
	Edition         string

	LastMod string
}

type Place struct {
	Name          string
	ShelvingIndex string
	Cipher        string
	Copies        int
	Available     int
}
