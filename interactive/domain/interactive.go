package domain

type Interactive struct {
	Biz        string
	BizId      int64 `json:"bizId,omitempty"`
	ReadCnt    int64 `json:"readCnt,omitempty"`
	LikeCnt    int64 `json:"likeCnt,omitempty"`
	CollectCnt int64 `json:"collectCnt,omitempty"`
	Liked      bool  `json:"liked,omitempty"`
	Collected  bool  `json:"collected,omitempty"`
}
