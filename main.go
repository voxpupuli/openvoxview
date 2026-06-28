package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	logger := cfg.GetLogger()
	slog.SetDefault(logger)

	slog.Info(fmt.Sprintf("OpenVox View - %s (%s)", VERSION, COMMIT))
	slog.Info(fmt.Sprintf("LISTEN: %s", cfg.Listen))
	slog.Info(fmt.Sprintf("PORT: %d", cfg.Port))
	slog.Info(fmt.Sprintf("PUPPETDB_ADDRESS: %s", cfg.GetPuppetDbAddress()))
	slog.Info(fmt.Sprintf("TRUSTED_PROXIES: %s", cfg.TrustedProxies))

	r := gin.New()
	r.Use(SlogMiddleware(logger))

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
				CaEnabled                         bool
				CaReadOnly                        bool
				UnreportedHours                   uint64
				StripPathPrefix                   string
				UiDefaultRefreshIntervalInSeconds uint
			}

			response := metaResponse{
				CaEnabled:                         caEnabled,
				CaReadOnly:                        cfg.PuppetCA.ReadOnly,
				UnreportedHours:                   cfg.UnreportedHours,
				StripPathPrefix:                   cfg.StripPathPrefix,
				UiDefaultRefreshIntervalInSeconds: cfg.UiDefaultRefreshIntervalInSeconds,
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

		ca.POST("status", caHandler.QueryCertificateStatuses)
		if !cfg.PuppetCA.ReadOnly {
			ca.POST("status/:name/sign", caHandler.SignCertificate)
			ca.POST("status/:name/revoke", caHandler.RevokeCertificate)
			ca.DELETE("status/:name", caHandler.CleanCertificate)
		}
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

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		logger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("body_size", c.Writer.Size()),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error", slog.String("error", err.Error()))
			}
		}
	}
}
