package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/model"
	"github.com/sebastianrakel/openvoxview/puppetdb"
)

type OpenVoxViewCollector struct {
	config    *config.Config
	pdbClient puppetdb.Client

	nodesFailedDesc         *prometheus.Desc
	nodesPendingDesc        *prometheus.Desc
	nodesChangedDesc        *prometheus.Desc
	nodesUnchangedDesc      *prometheus.Desc
	nodesUnreportedDesc     *prometheus.Desc
	populationDesc          *prometheus.Desc
	resourcesManagedDesc    *prometheus.Desc
	avgResourcesPerNodeDesc *prometheus.Desc
}

func NewOpenVoxViewCollector(cfg *config.Config, pdbClient puppetdb.Client) *OpenVoxViewCollector {
	return &OpenVoxViewCollector{
		config:    cfg,
		pdbClient: pdbClient,
		nodesFailedDesc: prometheus.NewDesc(
			"openvoxview_nodes_failed",
			"Number of nodes with latest report status failed.",
			nil,
			nil,
		),
		nodesPendingDesc: prometheus.NewDesc(
			"openvoxview_nodes_pending",
			"Number of nodes with latest report status pending.",
			nil,
			nil,
		),
		nodesChangedDesc: prometheus.NewDesc(
			"openvoxview_nodes_changed",
			"Number of nodes with latest report status changed.",
			nil,
			nil,
		),
		nodesUnchangedDesc: prometheus.NewDesc(
			"openvoxview_nodes_unchanged",
			"Number of nodes with latest report status unchanged.",
			nil,
			nil,
		),
		nodesUnreportedDesc: prometheus.NewDesc(
			"openvoxview_nodes_unreported",
			"Number of nodes that have not reported within the configured unreported duration.",
			nil,
			nil,
		),
		populationDesc: prometheus.NewDesc(
			"openvoxview_population",
			"Total number of nodes (population).",
			nil,
			nil,
		),
		resourcesManagedDesc: prometheus.NewDesc(
			"openvoxview_resources_managed",
			"Total number of resources managed.",
			nil,
			nil,
		),
		avgResourcesPerNodeDesc: prometheus.NewDesc(
			"openvoxview_avg_resources_per_node",
			"Average number of resources per node.",
			nil,
			nil,
		),
	}
}

func (c *OpenVoxViewCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nodesFailedDesc
	ch <- c.nodesPendingDesc
	ch <- c.nodesChangedDesc
	ch <- c.nodesUnchangedDesc
	ch <- c.nodesUnreportedDesc
	ch <- c.populationDesc
	ch <- c.resourcesManagedDesc
	ch <- c.avgResourcesPerNodeDesc
}

type countResult struct {
	Count float64 `json:"count"`
}

func queryPqlCount(client puppetdb.Client, pql string) (float64, error) {
	res, _, err := client.Query(pql)
	if err != nil {
		return 0, err
	}
	if len(res) == 0 {
		return 0, errors.New("empty response from puppetdb")
	}
	var cr countResult
	if err := json.Unmarshal(res[0], &cr); err != nil {
		return 0, err
	}
	return cr.Count, nil
}

func (c *OpenVoxViewCollector) Collect(ch chan<- prometheus.Metric) {
	var (
		nodes    []model.Node
		nodesErr error
		popCount float64
		popErr   error
		resCount float64
		resErr   error
		wg       sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		nodes, nodesErr = c.pdbClient.GetNodes(&puppetdb.PdbQuery{})
	}()

	go func() {
		defer wg.Done()
		popCount, popErr = queryPqlCount(c.pdbClient, "nodes [count()] { }")
	}()

	go func() {
		defer wg.Done()
		resCount, resErr = queryPqlCount(c.pdbClient, "resources [count()] { }")
	}()

	wg.Wait()

	if nodesErr != nil {
		slog.Error("failed to get nodes from puppetdb for prometheus metrics", "error", nodesErr)
		return
	}

	var failed, pending, changed, unchanged, unreported float64
	unreportedHours := uint64(3)
	if c.config != nil && c.config.UnreportedHours > 0 {
		unreportedHours = c.config.UnreportedHours
	}
	unreportedThreshold := time.Now().Add(-time.Duration(unreportedHours) * time.Hour)

	for _, node := range nodes {
		switch node.LatestReportStatus {
		case "failed":
			failed++
		case "pending":
			pending++
		case "changed":
			changed++
		case "unchanged":
			unchanged++
		}

		if node.ReportTimestamp == nil || node.ReportTimestamp.IsZero() || node.ReportTimestamp.Before(unreportedThreshold) {
			unreported++
		}
	}

	population := popCount
	if popErr != nil {
		slog.Warn("failed to query population count from puppetdb, falling back to node list length", "error", popErr)
		population = float64(len(nodes))
	}

	resources := resCount
	if resErr != nil {
		slog.Error("failed to query resources count from puppetdb for prometheus metrics", "error", resErr)
		resources = 0
	}

	var avgResourcesPerNode float64
	if population > 0 {
		avgResourcesPerNode = resources / population
	}

	ch <- prometheus.MustNewConstMetric(c.nodesFailedDesc, prometheus.GaugeValue, failed)
	ch <- prometheus.MustNewConstMetric(c.nodesPendingDesc, prometheus.GaugeValue, pending)
	ch <- prometheus.MustNewConstMetric(c.nodesChangedDesc, prometheus.GaugeValue, changed)
	ch <- prometheus.MustNewConstMetric(c.nodesUnchangedDesc, prometheus.GaugeValue, unchanged)
	ch <- prometheus.MustNewConstMetric(c.nodesUnreportedDesc, prometheus.GaugeValue, unreported)
	ch <- prometheus.MustNewConstMetric(c.populationDesc, prometheus.GaugeValue, population)
	ch <- prometheus.MustNewConstMetric(c.resourcesManagedDesc, prometheus.GaugeValue, resources)
	ch <- prometheus.MustNewConstMetric(c.avgResourcesPerNodeDesc, prometheus.GaugeValue, avgResourcesPerNode)
}

type MetricsHandler struct {
	collector *OpenVoxViewCollector
	registry  *prometheus.Registry
}

func NewMetricsHandler(cfg *config.Config, pdbClient puppetdb.Client) *MetricsHandler {
	collector := NewOpenVoxViewCollector(cfg, pdbClient)
	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return &MetricsHandler{
		collector: collector,
		registry:  reg,
	}
}

func (h *MetricsHandler) Handler() gin.HandlerFunc {
	promHandler := promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{})
	return gin.WrapH(promHandler)
}
