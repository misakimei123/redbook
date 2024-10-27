package web

type Page struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type ArticleVo struct {
	Id         int64  `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Abstract   string `json:"abstract,omitempty"`
	Content    string `json:"content,omitempty"`
	AuthorId   int64  `json:"authorId,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
	Status     uint8  `json:"status,omitempty"`
	Ctime      string `json:"ctime,omitempty"`
	Utime      string `json:"utime,omitempty"`
	ReadCnt    int64  `json:"readCnt,omitempty"`
	LikeCnt    int64  `json:"likeCnt,omitempty"`
	CollectCnt int64  `json:"collectCnt,omitempty"`
	Liked      bool   `json:"liked,omitempty"`
	Collected  bool   `json:"collected,omitempty"`
}
