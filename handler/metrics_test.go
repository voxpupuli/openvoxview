package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/model"
	"github.com/sebastianrakel/openvoxview/puppetdb"
)

type mockPuppetDbClient struct {
	nodes     []model.Node
	nodesErr  error
	queryFunc func(query string) ([]json.RawMessage, int, error)
}

func (m *mockPuppetDbClient) Query(query string) ([]json.RawMessage, int, error) {
	if m.queryFunc != nil {
		return m.queryFunc(query)
	}
	return nil, http.StatusOK, nil
}

func (m *mockPuppetDbClient) GetNodes(query *puppetdb.PdbQuery) ([]model.Node, error) {
	return m.nodes, m.nodesErr
}

func (m *mockPuppetDbClient) GetFacts(query *puppetdb.PdbQuery) ([]model.Fact, error) {
	return nil, nil
}

func (m *mockPuppetDbClient) GetFactNames() (json.RawMessage, error) {
	return nil, nil
}

func (m *mockPuppetDbClient) GetEventCounts(query *puppetdb.PdbQuery) ([]model.EventCount, error) {
	return nil, nil
}

func (m *mockPuppetDbClient) GetMetric(metricName string) (model.Metric, error) {
	return model.Metric{}, nil
}

func (m *mockPuppetDbClient) GetMetricList() (model.MetricList, error) {
	return model.MetricList{}, nil
}

func (m *mockPuppetDbClient) DeactivateNode(certname string) (*model.CommandResponse, error) {
	return nil, nil
}

func TestCollectorCollect(t *testing.T) {
	now := time.Now()
	recentTime := model.PuppetTime{Time: now.Add(-30 * time.Minute)}
	oldTime := model.PuppetTime{Time: now.Add(-5 * time.Hour)}

	nodes := []model.Node{
		{
			Name:               "node1.example.com",
			LatestReportStatus: "failed",
			ReportTimestamp:    &recentTime,
		},
		{
			Name:               "node2.example.com",
			LatestReportStatus: "failed",
			ReportTimestamp:    &oldTime, // unreported
		},
		{
			Name:               "node3.example.com",
			LatestReportStatus: "pending",
			ReportTimestamp:    &recentTime,
		},
		{
			Name:               "node4.example.com",
			LatestReportStatus: "changed",
			ReportTimestamp:    &recentTime,
		},
		{
			Name:               "node5.example.com",
			LatestReportStatus: "changed",
			ReportTimestamp:    &recentTime,
		},
		{
			Name:               "node6.example.com",
			LatestReportStatus: "changed",
			ReportTimestamp:    nil, // unreported
		},
		{
			Name:               "node7.example.com",
			LatestReportStatus: "unchanged",
			ReportTimestamp:    &recentTime,
		},
		{
			Name:               "node8.example.com",
			LatestReportStatus: "unchanged",
			ReportTimestamp:    &recentTime,
		},
		{
			Name:               "node9.example.com",
			LatestReportStatus: "unchanged",
			ReportTimestamp:    &recentTime,
		},
		{
			Name:               "node10.example.com",
			LatestReportStatus: "unchanged",
			ReportTimestamp:    &oldTime, // unreported
		},
	}

	mockPdb := &mockPuppetDbClient{
		nodes: nodes,
		queryFunc: func(query string) ([]json.RawMessage, int, error) {
			if strings.Contains(query, "nodes") {
				raw := json.RawMessage(`{"count": 10}`)
				return []json.RawMessage{raw}, http.StatusOK, nil
			}
			if strings.Contains(query, "resources") {
				raw := json.RawMessage(`{"count": 500}`)
				return []json.RawMessage{raw}, http.StatusOK, nil
			}
			return nil, http.StatusOK, nil
		},
	}

	cfg := &config.Config{
		UnreportedHours: 3,
	}

	collector := NewOpenVoxViewCollector(cfg, mockPdb)
	reg := prometheus.NewRegistry()
	if err := reg.Register(collector); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	expectedMetrics := `
# HELP openvoxview_avg_resources_per_node Average number of resources per node.
# TYPE openvoxview_avg_resources_per_node gauge
openvoxview_avg_resources_per_node 50
# HELP openvoxview_nodes_changed Number of nodes with latest report status changed.
# TYPE openvoxview_nodes_changed gauge
openvoxview_nodes_changed 3
# HELP openvoxview_nodes_failed Number of nodes with latest report status failed.
# TYPE openvoxview_nodes_failed gauge
openvoxview_nodes_failed 2
# HELP openvoxview_nodes_pending Number of nodes with latest report status pending.
# TYPE openvoxview_nodes_pending gauge
openvoxview_nodes_pending 1
# HELP openvoxview_nodes_unchanged Number of nodes with latest report status unchanged.
# TYPE openvoxview_nodes_unchanged gauge
openvoxview_nodes_unchanged 4
# HELP openvoxview_nodes_unreported Number of nodes that have not reported within the configured unreported duration.
# TYPE openvoxview_nodes_unreported gauge
openvoxview_nodes_unreported 3
# HELP openvoxview_population Total number of nodes (population).
# TYPE openvoxview_population gauge
openvoxview_population 10
# HELP openvoxview_resources_managed Total number of resources managed.
# TYPE openvoxview_resources_managed gauge
openvoxview_resources_managed 500
`

	metricNames := []string{
		"openvoxview_nodes_failed",
		"openvoxview_nodes_pending",
		"openvoxview_nodes_changed",
		"openvoxview_nodes_unchanged",
		"openvoxview_nodes_unreported",
		"openvoxview_population",
		"openvoxview_resources_managed",
		"openvoxview_avg_resources_per_node",
	}

	if err := testutil.GatherAndCompare(reg, strings.NewReader(expectedMetrics), metricNames...); err != nil {
		t.Errorf("gathered metrics did not match expected: %v", err)
	}
}

func TestCollectorWithPdbErrors(t *testing.T) {
	mockPdb := &mockPuppetDbClient{
		nodesErr: errors.New("connection refused"),
	}

	cfg := &config.Config{
		UnreportedHours: 3,
	}

	collector := NewOpenVoxViewCollector(cfg, mockPdb)
	reg := prometheus.NewRegistry()
	if err := reg.Register(collector); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	// When nodes query errors, no metrics should be collected
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("unexpected error during gather: %v", err)
	}

	for _, mf := range mfs {
		if strings.HasPrefix(mf.GetName(), "openvoxview_") {
			t.Errorf("expected no openvoxview metrics on error, got %s", mf.GetName())
		}
	}
}

func TestCollectorPopulationFallback(t *testing.T) {
	nodes := []model.Node{
		{Name: "n1", LatestReportStatus: "unchanged"},
		{Name: "n2", LatestReportStatus: "failed"},
	}

	mockPdb := &mockPuppetDbClient{
		nodes: nodes,
		queryFunc: func(query string) ([]json.RawMessage, int, error) {
			// Fail both population and resources queries
			return nil, http.StatusInternalServerError, errors.New("pql failure")
		},
	}

	cfg := &config.Config{
		UnreportedHours: 3,
	}

	collector := NewOpenVoxViewCollector(cfg, mockPdb)
	reg := prometheus.NewRegistry()
	if err := reg.Register(collector); err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	expectedMetrics := `
# HELP openvoxview_avg_resources_per_node Average number of resources per node.
# TYPE openvoxview_avg_resources_per_node gauge
openvoxview_avg_resources_per_node 0
# HELP openvoxview_nodes_changed Number of nodes with latest report status changed.
# TYPE openvoxview_nodes_changed gauge
openvoxview_nodes_changed 0
# HELP openvoxview_nodes_failed Number of nodes with latest report status failed.
# TYPE openvoxview_nodes_failed gauge
openvoxview_nodes_failed 1
# HELP openvoxview_nodes_pending Number of nodes with latest report status pending.
# TYPE openvoxview_nodes_pending gauge
openvoxview_nodes_pending 0
# HELP openvoxview_nodes_unchanged Number of nodes with latest report status unchanged.
# TYPE openvoxview_nodes_unchanged gauge
openvoxview_nodes_unchanged 1
# HELP openvoxview_nodes_unreported Number of nodes that have not reported within the configured unreported duration.
# TYPE openvoxview_nodes_unreported gauge
openvoxview_nodes_unreported 2
# HELP openvoxview_population Total number of nodes (population).
# TYPE openvoxview_population gauge
openvoxview_population 2
# HELP openvoxview_resources_managed Total number of resources managed.
# TYPE openvoxview_resources_managed gauge
openvoxview_resources_managed 0
`

	metricNames := []string{
		"openvoxview_nodes_failed",
		"openvoxview_nodes_pending",
		"openvoxview_nodes_changed",
		"openvoxview_nodes_unchanged",
		"openvoxview_nodes_unreported",
		"openvoxview_population",
		"openvoxview_resources_managed",
		"openvoxview_avg_resources_per_node",
	}

	if err := testutil.GatherAndCompare(reg, strings.NewReader(expectedMetrics), metricNames...); err != nil {
		t.Errorf("gathered metrics did not match expected: %v", err)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPdb := &mockPuppetDbClient{
		nodes: []model.Node{
			{Name: "n1", LatestReportStatus: "failed"},
		},
		queryFunc: func(query string) ([]json.RawMessage, int, error) {
			if strings.Contains(query, "nodes") {
				return []json.RawMessage{json.RawMessage(`{"count": 1}`)}, http.StatusOK, nil
			}
			if strings.Contains(query, "resources") {
				return []json.RawMessage{json.RawMessage(`{"count": 10}`)}, http.StatusOK, nil
			}
			return nil, http.StatusOK, nil
		},
	}

	cfg := &config.Config{
		UnreportedHours: 3,
	}

	handler := NewMetricsHandler(cfg, mockPdb)

	router := gin.New()
	router.GET("/metrics", handler.Handler())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(body)
	expectedSubstrings := []string{
		"openvoxview_nodes_failed 1",
		"openvoxview_nodes_pending 0",
		"openvoxview_nodes_changed 0",
		"openvoxview_nodes_unchanged 0",
		"openvoxview_nodes_unreported 1",
		"openvoxview_population 1",
		"openvoxview_resources_managed 10",
		"openvoxview_avg_resources_per_node 10",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(bodyStr, sub) {
			t.Errorf("expected body to contain %q, but it didn't.\nBody:\n%s", sub, bodyStr)
		}
	}
}
