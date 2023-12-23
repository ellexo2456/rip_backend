package ds

const AuthKey = "auth"

type ArchiveResponse struct {
	Archived bool   `json:"archived"`
	AuthKey  string `json:"authKey"`
}
