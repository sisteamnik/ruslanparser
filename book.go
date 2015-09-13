package ruslanparser

type Book struct {
	Id       int64    `json:"id"`
	SourceId string   `json:"source_id"`
	Title    string   `json:"title"`
	Author   string   `json:"author"`
	Series   string   `json:"series"`
	Tags     []string `db:"-" json:"tags"`
	Genre    string   `json:"genre"`

	Places []Place `db:"-" json:"places"`

	PageNum int `json:"pagenum"`

	City            string `json:"city"`
	Publishing      string `json:"publishing"`
	PublicationYear int    `json:"publication_year"`
	Edition         string `json:"edition"`

	Fields map[string]map[string]string `json:"fields"`

	LastMod string `json:"lastmod"`
}

type Place struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	ShelvingIndex string `json:"shelving_index"`
	Cipher        string `json:"cipher"`
	Copies        int    `json:"copies"`
	Available     int    `json:"available"`
}

type Places struct {
	Id     int64 `json:"id"`
	BookId int64 `json:"book_id"`
}

type Tag struct {
	Id    int64
	Title string
}

type Tags struct {
	Id     int64
	BookId int64
}
