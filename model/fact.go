package model

type Fact struct {
	Certname    string `json:"certname"`
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Value       any    `json:"value"`
}
