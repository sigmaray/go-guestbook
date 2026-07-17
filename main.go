package main

import (
	"embed"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"go-guestbook/cli"
	"go-guestbook/config"
	"go-guestbook/database"
	"go-guestbook/handlers"
	"go-guestbook/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// main starts the guestbook CLI or HTTP server depending on command-line arguments.
func main() {
	database.LoadEnv()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if len(os.Args) < 2 {
		cli.PrintUsage()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	switch os.Args[1] {
	case "s", "server":
		runServer(cfg)
	case "migrate":
		if err := database.RunMigrations(cfg, embedMigrations); err != nil {
			log.Fatal().Err(err).Msg("migration failed")
		}
	default:
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to database")
		}
		cli.Run(db, os.Args[1:])
	}
}

// runServer configures Gin, registers routes, and starts the HTTP listener.
// cfg is the application configuration loaded from the environment.
func runServer(cfg *config.Config) {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	middleware.ConfigureRateLimits(middleware.RateLimitConfig{
		LoginEnabled:   cfg.LoginRateLimitEnabled,
		LoginMax:       cfg.LoginRateLimitMaxAttempts,
		LoginWindow:    cfg.LoginRateLimitWindow,
		MessageEnabled: cfg.MessageRateLimitEnabled,
		MessageMax:     cfg.MessageRateLimitMaxAttempts,
		MessageWindow:  cfg.MessageRateLimitWindow,
	})

	h := handlers.NewHandler(db, log.Logger, cfg)

	r.GET("/health", h.Health)
	r.HEAD("/health", h.Health)

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   cfg.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("guestbook_session", store))

	loadHTMLTemplates(r)

	r.Static("/static", "./static")
	r.StaticFile("/robots.txt", "robots.txt")

	r.GET("/", h.Index)
	r.HEAD("/", h.Health)
	r.POST("/messages", h.CreateMessage)
	r.GET("/login", h.LoginPage)
	r.POST("/login", h.Login)

	if cfg.TestAPIEnabled {
		testAPI := r.Group("/api/test")
		{
			testAPI.POST("/truncate", h.TruncateTable)
			testAPI.POST("/entities", h.CreateEntity)
			testAPI.POST("/sql", h.ExecuteSQL)
		}
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired())
	{
		admin.GET("/", h.AdminMessagesRedirect)
		admin.GET("/messages", h.AdminMessagesList)
		admin.GET("/messages/new", h.NewMessagePage)
		admin.POST("/messages", h.CreateAdminMessage)
		admin.GET("/messages/:id", h.ShowMessagePage)
		admin.GET("/messages/:id/edit", h.EditMessagePage)
		admin.POST("/messages/:id", h.UpdateMessage)
		admin.POST("/messages/:id/delete", h.DeleteMessage)

		admin.GET("/users", h.UsersList)
		admin.GET("/users/new", h.NewUserPage)
		admin.POST("/users", h.CreateUser)
		admin.GET("/users/:id/edit", h.EditUserPage)
		admin.POST("/users/:id", h.UpdateUser)
		admin.POST("/users/:id/delete", h.DeleteUser)

		tools := admin.Group("/tools")
		tools.Use(middleware.DevelopmentOnly(cfg))
		{
			tools.GET("/", h.ToolsPage)
			tools.POST("/clear-table", h.ToolsClearTable)
			tools.POST("/execute-sql", h.ToolsExecuteSQL)
			tools.POST("/seed-messages", h.ToolsSeedMessages)
		}

		admin.POST("/logout", h.Logout)
	}

	port := cfg.HTTPPort
	if port == 0 {
		port = 8084
	}
	addr := ":" + strconv.Itoa(port)
	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Str("addr", addr).Msg("failed to start server")
	}
}

// loadHTMLTemplates parses admin and public HTML templates into the Gin engine.
// r is the Gin engine that will serve the parsed templates.
func loadHTMLTemplates(r *gin.Engine) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
	}

	patterns := []string{
		"templates/admin/*.html",
		"templates/admin/*/*.html",
		"templates/public/*.html",
	}

	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Fatal().Err(err).Str("pattern", pattern).Msg("failed to glob templates")
		}
		files = append(files, matches...)
	}

	if len(files) == 0 {
		log.Fatal().Msg("no HTML templates found")
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(files...))
	r.SetHTMLTemplate(tmpl)
}
