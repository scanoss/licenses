package dto

type ComponentRequestDTO struct {
	Purl        string `json:"purl"`
	Requirement string `json:"requirement,omitempty"`
}
