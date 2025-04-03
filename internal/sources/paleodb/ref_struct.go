package paleodb

// Author represents the author structure
type Author struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// Identifier represents the identifier structure
type Identifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Reference struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Year       string     `json:"year"`
	Author     []Author   `json:"author"`
	Journal    string     `json:"journal,omitempty"`
	Volume     string     `json:"volume,omitempty"`
	Number     string     `json:"number,omitempty"`
	Pages      string     `json:"pages,omitempty"`
	Language   string     `json:"language,omitempty"`
	Identifier Identifier `json:"identifier,omitempty"`
	Publisher  string     `json:"publisher,omitempty"`
	ISBN       string     `json:"isbn,omitempty"`
}

type References struct {
	ElapsedTime float64     `json:"elapsed_time"`
	Records     []Reference `json:"records"`
}
