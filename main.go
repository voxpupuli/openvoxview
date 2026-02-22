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

var (
	VERSION = "0.1.0"
	COMMIT  = "dirty"
)

func main() {
	if config.PrintVersion(VERSION) {
		return
	}
	log.Printf("OpenVox View - %s (%s)", VERSION, COMMIT)
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	log.Printf("LISTEN: %s", cfg.Listen)
	log.Printf("PORT: %d", cfg.Port)
	log.Printf("PUPPETDB_ADDRESS: %s", cfg.GetPuppetDbAddress())
	log.Printf("TRUSTED_PROXIES: %#v", cfg.TrustedProxies)

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

	if len(cfg.TrustedProxies) > 0 {
		r.SetTrustedProxies(cfg.TrustedProxies)
	}

	caEnabled := cfg.PuppetCA.Host != ""

	pdbHandler := handler.NewPdbHandler(cfg)
	viewHandler := handler.NewViewHandler(cfg)

	api := r.Group("/api/v1/")
	{
		api.GET("meta", func(c *gin.Context) {
			type metaResponse struct {
				CaEnabled       bool
				CaReadOnly      bool
				UnreportedHours uint64
			}

			response := metaResponse{
				CaEnabled:       caEnabled,
				CaReadOnly:      cfg.PuppetCA.ReadOnly,
				UnreportedHours: cfg.UnreportedHours,
			}

			c.JSON(http.StatusOK, handler.NewSuccessResponse(response))
		})
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
			view.GET("predefined/:viewName/meta", viewHandler.PredefinedViewsMeta)
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

	if caEnabled {
		caHandler := handler.NewCaHandler(cfg)
		ca := api.Group("ca")
		caHandler.RegisterRoutes(ca)
	}

	r.Run(fmt.Sprintf("%s:%d", cfg.Listen, cfg.Port))
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
