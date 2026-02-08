package model

import "time"

type Node struct {
	// Fields in OpenVoxDB nodes response format:
	// https://github.com/OpenVoxProject/openvoxdb/blob/8.12.1/documentation/api/query/v4/nodes.markdown#response-format
	Name                    string     `json:"certname"`
	Deactivated             *time.Time `json:"deactivated"`
	Expired                 *time.Time `json:"expired"`
	CatalogTimestamp        *time.Time `json:"catalog_timestamp"`
	FactsTimestamp          *time.Time `json:"facts_timestamp"`
	ReportTimestamp         *time.Time `json:"report_timestamp"`
	CatalogEnvironment      *string    `json:"catalog_environment"`
	FactsEnvironment        *string    `json:"facts_environment"`
	ReportEnvironment       *string    `json:"report_environment"`
	LatestReportStatus      string     `json:"latest_report_status"`
	LatestReportNoop        bool       `json:"latest_report_noop"`
	LatestReportNoopPending bool       `json:"latest_report_noop_pending"`
	LatestReportHash        string     `json:"latest_report_hash"`
	LatestReportJobId       string     `json:"latest_report_job_id"`

	// Fields not in the docs linked above, but are in the API response:
	CachedCatalogStatus          *string `json:"cached_catalog_status"`
	LatestReportCorrectiveChange *bool   `json:"latest_report_corrective_change"`

	// Additional fields for our use, not in the OpenVoxDB API response:
	Events EventCount `json:"events"`
}

func NodeFromData(nodeData map[string]interface{}, eventData interface{}) Node {
	return Node{
		Name: nodeData["certname"].(string),
	}
}
