package model

type Language struct {
	Bytes int64 `json:"bytes"`
}

type Repository struct {
	FullName   string              `json:"full_name"`
	Owner      string              `json:"owner"`
	Repository string              `json:"repository"`
	Languages  map[string]Language `json:"languages"`
}
