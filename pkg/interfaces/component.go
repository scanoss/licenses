package interfaces

type Component struct {
	Purl        string `json:"purl"`
	Requirement string `json:"requirement,omitempty"`
}
