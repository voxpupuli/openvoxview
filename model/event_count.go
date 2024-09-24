package model

type EventCount struct {
	Failures    int    `json:"failures"`
	Skips       int    `json:"skips"`
	Successes   int    `json:"successes"`
	Noops       int    `json:"noops"`
	SubjectType string `json:"subject_type"`
	Subject     struct {
		Title string `json:"title"`
	} `json:"subject"`
}
