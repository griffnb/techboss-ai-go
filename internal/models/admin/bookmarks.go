package admin

type Bookmarks struct {
	Reports []*Bookmark `json:"reports,omitempty"`
	Pages   []*Bookmark `json:"pages,omitempty"`
}

type Bookmark struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}
