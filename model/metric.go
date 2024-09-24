package model

type Metric struct {
	Request struct {
		Mbean string `json:"mbean"`
		Type  string `json:"type"`
	} `json:"request"`
	Value struct {
		Value any `json:"Value"`
	} `json:"value"`
	Timestamp uint `json:"timestamp"`
	Status    uint `json:"status"`
}

type MetricInfoAttr struct {
	Rw   bool
	Type string
	Desc string
}

type MetricInfoOp struct {
	Args []any
	Ret  string
	Desc string
}

type MetricInfo struct {
	Op    map[string]MetricInfoOp
	Attr  map[string]MetricInfoAttr
	Desc  string
	Class string
}

type MetricList struct {
	Request struct {
		Type string
	}
	Value map[string]MetricInfo
}
