package util

type PageResult struct {
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
	Count    int64       `json:"count"`
	Data     interface{} `json:"data"`
}
