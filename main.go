package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/db"
	"github.com/sebastianrakel/openvoxview/handler"
	"github.com/sebastianrakel/openvoxview/middleware"
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

	// Handle --create-admin before starting the server
	if config.CreateAdmin() {
		runCreateAdmin(cfg)
		return
	}

	log.Printf("LISTEN: %s", cfg.Listen)
	log.Printf("PORT: %d", cfg.Port)
	log.Printf("PUPPETDB_ADDRESS: %s", cfg.GetPuppetDbAddress())
	log.Printf("TRUSTED_PROXIES: %#v", cfg.TrustedProxies)

	// Initialize auth database if auth is enabled
	var database *db.Database
	if cfg.Auth.Enabled {
		if cfg.Auth.JwtSecret == "" {
			cfg.Auth.JwtSecret = generateRandomSecret()
			log.Printf("WARNING: No jwt_secret configured. A random secret was generated. Tokens will not survive restarts. Set auth.jwt_secret in your config.")
		}
		if len(cfg.Auth.JwtSecret) < 32 {
			log.Printf("WARNING: jwt_secret is shorter than 32 characters. This is insecure for production use.")
		}

		database, err = db.Open(cfg.Auth.DbPath)
		if err != nil {
			log.Fatalf("Failed to open auth database: %v", err)
		}
		defer database.Close()

		count, _ := database.UserCount()
		if count == 0 {
			log.Printf("WARNING: Auth is enabled but no users exist. Use --create-admin to create the first user.")
		}

		// Start periodic token cleanup
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				database.CleanupExpiredTokens()
			}
		}()

		log.Printf("AUTH: enabled (db: %s)", cfg.Auth.DbPath)
	} else {
		log.Printf("AUTH: disabled")
	}

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

	// Public auth endpoints (no JWT required)
	if cfg.Auth.Enabled {
		authHandler := handler.NewAuthHandler(cfg, database)
		r.POST("/api/v1/auth/login", authHandler.Login)
		r.POST("/api/v1/auth/refresh", authHandler.Refresh)
	}

	// Public endpoints (no JWT required)
	r.GET("/api/v1/version", func(c *gin.Context) {
		type versionResponse struct {
			Version string
		}
		c.JSON(http.StatusOK, handler.NewSuccessResponse(versionResponse{Version: VERSION}))
	})

	r.GET("/api/v1/meta", func(c *gin.Context) {
		type metaResponse struct {
			CaEnabled       bool
			CaReadOnly      bool
			UnreportedHours uint64
			StripPathPrefix string
			AuthEnabled     bool
		}

		response := metaResponse{
			CaEnabled:       caEnabled,
			CaReadOnly:      cfg.PuppetCA.ReadOnly,
			UnreportedHours: cfg.UnreportedHours,
			StripPathPrefix: cfg.StripPathPrefix,
			AuthEnabled:     cfg.Auth.Enabled,
		}

		c.JSON(http.StatusOK, handler.NewSuccessResponse(response))
	})

	pdbHandler := handler.NewPdbHandler(cfg)
	viewHandler := handler.NewViewHandler(cfg)

	api := r.Group("/api/v1/")
	api.Use(middleware.JWTAuthMiddleware(cfg))
	{

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

		// Auth management endpoints (require auth)
		if cfg.Auth.Enabled {
			authHandler := handler.NewAuthHandler(cfg, database)
			auth := api.Group("auth")
			{
				auth.POST("logout", authHandler.Logout)
				auth.GET("me", authHandler.Me)
				auth.GET("users", authHandler.ListUsers)
				auth.POST("users", authHandler.CreateUser)
				auth.PUT("users/:id", authHandler.UpdateUser)
				auth.DELETE("users/:id", authHandler.DeleteUser)
			}
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

func runCreateAdmin(cfg *config.Config) {
	database, err := db.Open(cfg.Auth.DbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		log.Fatal("Username cannot be empty")
	}

	fmt.Print("Email (optional): ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Display Name (optional): ")
	displayName, _ := reader.ReadString('\n')
	displayName = strings.TrimSpace(displayName)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		log.Fatal("Password must be at least 8 characters")
	}

	fmt.Print("Confirm Password: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)
	if password != confirm {
		log.Fatal("Passwords do not match")
	}

	user, err := database.CreateUser(username, email, displayName, password)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Admin user created: %s (id: %d)\n", user.Username, user.ID)
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
