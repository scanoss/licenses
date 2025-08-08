package dto

type ComponentRequestDTO struct {
	Purl         string `json:"purl"`
	Requirement  string `json:"requirement,omitempty"`
	OriginalPurl string `json:"-"` // Internal field - preserve original PURL format when split from purl@version
	WasSplit     bool   `json:"-"` // Internal field - tracks if this component was split from purl@version
}
