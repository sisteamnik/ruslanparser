package ruslanparser

type Book struct {
	Id       int64
	SourceId string
	Title    string
	Author   string
	Series   string
	Tags     []string `db:"-"`
	Genre    string

	Places []Place `db:"-"`

	PageNum int

	City            string
	Publishing      string
	PublicationYear int
	Edition         string

	LastMod string
}

type Place struct {
	Id            int64
	Name          string
	ShelvingIndex string
	Cipher        string
	Copies        int
	Available     int
}

type Places struct {
	Id     int64
	BookId int64
}

type Tag struct {
	Id    int64
	Title string
}

type Tags struct {
	Id     int64
	BookId int64
}
