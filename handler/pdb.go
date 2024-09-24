package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/puppetdb"
)

type QueryRequest struct {
	Query         string
	SaveInHistory bool
}

type QueryResult struct {
	Data                 []json.RawMessage
	Error                error
	Success              bool
	ExecutedOn           time.Time
	ExecutionTimeInMilli int64
	Count                int
}

type PqlHistoryEntry struct {
	Query  QueryRequest
	Result QueryResult
}

type PdbHandler struct {
	QueryHistory []PqlHistoryEntry
	config       *config.Config
}

func NewPdbHandler(config *config.Config) *PdbHandler {
	return &PdbHandler{
		QueryHistory: []PqlHistoryEntry{},
		config:       config,
	}
}

func (h *PdbHandler) PdbExecuteQuery(c *gin.Context) {
	var queryRequest QueryRequest
	c.BindJSON(&queryRequest)

	dbClient := puppetdb.NewClient()
	log.Printf("Executing Query: %s", queryRequest.Query)

	historyEntry := PqlHistoryEntry{
		Query: queryRequest,
	}

	start := time.Now()
	res, err := dbClient.Query(queryRequest.Query)
	end := time.Now()

	duration := end.Sub(start).Milliseconds()

	queryResult := QueryResult{
		Data:                 res,
		Error:                err,
		Success:              err != nil,
		ExecutedOn:           time.Now(),
		ExecutionTimeInMilli: duration,
		Count:                len(res),
	}

	historyEntry.Result = queryResult

	if queryRequest.SaveInHistory {
		h.QueryHistory = append(h.QueryHistory, historyEntry)
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(queryResult))
}

func (h *PdbHandler) PdbQueryHistory(c *gin.Context) {
	c.JSON(http.StatusOK, NewSuccessResponse(h.QueryHistory))
}

func (h *PdbHandler) PdbQueryPredefined(c *gin.Context) {
	c.JSON(http.StatusOK, NewSuccessResponse(h.config.PqlQueries))
}

func (h *PdbHandler) PdbGetFactNames(c *gin.Context) {
	dbClient := puppetdb.NewClient()

	res, err := dbClient.GetFactNames()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(res))
}

func (h *PdbHandler) PdbGetEventCounts(c *gin.Context) {
	var query puppetdb.PdbQuery
	err := c.BindJSON(&query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	dbClient := puppetdb.NewClient()

	res, err := dbClient.GetEventCounts(&query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(res))
}
