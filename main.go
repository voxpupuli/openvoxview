package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/handler"
)

const (
	VERSION = "0.1.0"
	COMMIT  = "dirty"
)

func main() {
	log.Printf("OpenVox View - %s (%s)", VERSION, COMMIT)

	config, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	log.Printf("LISTEN: %s", config.Listen)
	log.Printf("PORT: %d", config.Port)
	log.Printf("PUPPET_SERVER_ADDRESS: %s", config.GetPuppetDbAddress())

	r := gin.Default()

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, "/ui/?#/")
	})

	uiFSSub, _ := fs.Sub(uiFS, "ui/dist/spa")
	r.StaticFS("ui", http.FS(uiFSSub))

	r.Use(AllowCORS)

	pdbHandler := handler.NewPdbHandler(config)
	viewHandler := handler.NewViewHandler(config)

	api := r.Group("/api/v1/")
	{
		api.GET("version", func(c *gin.Context) {
			type versionResponse struct {
				Version string
			}

			response := versionResponse{
				Version: VERSION,
			}

			c.JSON(http.StatusOK, handler.NewSuccessResponse(response))
		})
		view := api.Group("view")
		{
			view.GET("node_overview", viewHandler.NodesOverview)
			view.GET("metrics", viewHandler.Metrics)
			view.GET("predefined", viewHandler.PredefinedViews)
			view.GET("predefined/:viewName", viewHandler.PredefinedViewsResult)
		}

		pdb := api.Group("pdb")
		{
			pdb.POST("query", pdbHandler.PdbExecuteQuery)
			pdb.GET("query/history", pdbHandler.PdbQueryHistory)
			pdb.GET("query/predefined", pdbHandler.PdbQueryPredefined)
			pdb.GET("fact-names", pdbHandler.PdbGetFactNames)
			pdb.POST("event-counts", pdbHandler.PdbGetEventCounts)
		}
	}

	r.Run(fmt.Sprintf("%s:%d", config.Listen, config.Port))
}

func AllowCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	c.Header("Access-Control-Allow-Headers", "Authorization, *")

	if c.Request.Method == http.MethodOptions {
		c.Status(http.StatusNoContent)
		return
	}

	c.Next()
}
