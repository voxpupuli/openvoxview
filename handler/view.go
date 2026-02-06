package handler

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/model"
	"github.com/sebastianrakel/openvoxview/puppetdb"
)

type ViewHandler struct {
	config *config.Config
}

func NewViewHandler(config *config.Config) *ViewHandler {
	return &ViewHandler{
		config: config,
	}
}

type NodesOverviewQuery struct {
	Environment string   `form:"environment"`
	Status      []string `form:"status"`
}

func (n *NodesOverviewQuery) HasEnvironment() bool {
	return n.Environment != "*" && n.Environment != ""
}

func (h *ViewHandler) NodesOverview(c *gin.Context) {
	var nodesOverviewQuery NodesOverviewQuery
	err := c.BindQuery(&nodesOverviewQuery)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	dbClient := puppetdb.NewClient()

	eventCountsQuery := puppetdb.PdbQuery{
		Query: []any{
			"=",
			"latest_report?",
			true,
		},
		SummarizeBy: "certname",
	}

	eventCounts, err := dbClient.GetEventCounts(&eventCountsQuery)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	nodesQuery := &puppetdb.PdbQuery{
		Query: []any{},
	}

	if nodesOverviewQuery.HasEnvironment() {
		nodesQuery.Query = append(nodesQuery.Query,
			"and",
			[]any{
				"=",
				"catalog_environment",
				nodesOverviewQuery.Environment,
			})
	}

	if len(nodesOverviewQuery.Status) > 0 {
		orQuery := []any{
			"or",
		}

		for _, status := range nodesOverviewQuery.Status {
			orQuery = append(orQuery, []any{"=", "latest_report_status", status})
		}

		nodesQuery.Query = append(nodesQuery.Query, orQuery)
	}

	if len(nodesQuery.Query) == 0 {
		nodesQuery = nil
	}

	nodes, err := dbClient.GetNodes(nodesQuery)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	unreportedDuration := time.Duration(h.config.UnreportedHours) * time.Hour
	unreportedTime := time.Now().Add(-unreportedDuration)

	for i, node := range nodes {
		eventIndex := slices.IndexFunc(eventCounts, func(n model.EventCount) bool {
			return n.Subject.Title == node.Name
		})

		if eventIndex >= 0 {
			nodes[i].Events = eventCounts[eventIndex]
		}

		nodes[i].Unreported = (node.ReportTimestamp == nil || node.ReportTimestamp.Before(unreportedTime))
	}

	c.JSON(http.StatusOK, NewSuccessResponse(nodes))
}

func (h *ViewHandler) Metrics(c *gin.Context) {
	environment := c.Query("environment")

	dbClient := puppetdb.NewClient()

	if environment == "" || environment == "*" {
		dbClient.GetMetricList()
	} else {

	}
}

func (h *ViewHandler) PredefinedViews(c *gin.Context) {
	views := h.config.Views
	if views == nil {
		views = []model.View{}
	}

	c.JSON(http.StatusOK, NewSuccessResponse(views))
}

func (h *ViewHandler) PredefinedViewsResult(c *gin.Context) {
	viewName := c.Param("viewName")

	if viewName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(errors.New("no view name")))
		return
	}

	i := slices.IndexFunc(h.config.Views, func(n model.View) bool {
		return n.Name == viewName
	})

	if i < 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(errors.New("view does not exists")))
		return
	}

	predefinedView := h.config.Views[i]
	orQuery := []any{
		"or",
	}

	for _, fact := range predefinedView.Facts {
		root_fact := fact.Fact

		if strings.Contains(fact.Fact, ".") {
			sections := strings.Split(fact.Fact, ".")
			root_fact = sections[0]
		}

		orQuery = append(orQuery, []any{"=", "name", root_fact})
	}

	factsQuery := puppetdb.PdbQuery{
		Query: []any{
			"and",
			orQuery,
		},
	}

	dbClient := puppetdb.NewClient()
	facts, err := dbClient.GetFacts(&factsQuery)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	mapped := map[string]map[string]any{}
	for _, fact := range facts {
		if _, exists := mapped[fact.Certname]; !exists {
			mapped[fact.Certname] = map[string]any{}
		}

		mapped[fact.Certname][fact.Name] = fact.Value
	}

	flattend := []map[string]any{}
	for _, value := range mapped {
		flattend = append(flattend, value)
	}

	result := model.ViewResult{
		View: predefinedView,
		Data: flattend,
	}

	c.JSON(http.StatusOK, NewSuccessResponse(result))
}

func (h *ViewHandler) PredefinedViewsMeta(c *gin.Context) {
	viewName := c.Param("viewName")

	if viewName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(errors.New("no view name")))
		return
	}

	i := slices.IndexFunc(h.config.Views, func(n model.View) bool {
		return n.Name == viewName
	})

	if i < 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(errors.New("view does not exists")))
		return
	}

	predefinedView := h.config.Views[i]
	c.JSON(http.StatusOK, NewSuccessResponse(predefinedView))
}
