package model

type CaseStatusCount struct {
	Status string `json:"status,omitempty"`
	Count  int    `json:"count,omitempty"`
}

type CaseStatusItem struct {
	Status string `json:"status,omitempty"`
	Label  string `json:"label,omitempty"`
	Count  int    `json:"count,omitempty"`
	Target int    `json:"target,omitempty"`
	Color  string `json:"color,omitempty"`
	Helper string `json:"helper,omitempty"`
}
