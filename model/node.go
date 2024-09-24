package model

import "time"

type Node struct {
	Name                string     `json:"certname"`
	Deactivated         time.Time  `json:"deactivated"`
	Expired             bool       `json:"expired"`
	ReportTimestamp     time.Time  `json:"report_timestamp"`
	CatalogTimestamp    time.Time  `json:"catalog_timestamp"`
	FactsTimestamp      time.Time  `json:"facts_timestamp"`
	LatestReportStatus  string     `json:"latest_report_status"`
	Unreported          string     `json:"unreported"`
	ReportEnvironment   string     `json:"report_environment"`
	CatalogEnvironment  string     `json:"catalog_environment"`
	FactsEnvironment    string     `json:"facts_environment"`
	LatestReportHash    string     `json:"latest_report_hash"`
	CachedCatalogStatus string     `json:"cached_catalog_status"`
	Events              EventCount `json:"events"`
}

func NodeFromData(nodeData map[string]interface{}, eventData interface{}) Node {
	return Node{
		Name: nodeData["certname"].(string),
	}
}
