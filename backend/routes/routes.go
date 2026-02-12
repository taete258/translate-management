package routes

import (
	"translate-management/cache"
	"translate-management/config"
	"translate-management/handlers"
	"translate-management/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(app *fiber.App, db *pgxpool.Pool, rdb *cache.RedisClient, cfg *config.Config) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	projectHandler := handlers.NewProjectHandler(db)
	languageHandler := handlers.NewLanguageHandler(db)
	keyHandler := handlers.NewKeyHandler(db)
	translationHandler := handlers.NewTranslationHandler(db)
	apiKeyHandler := handlers.NewAPIKeyHandler(db)
	cacheHandler := handlers.NewCacheHandler(db, rdb)
	exportHandler := handlers.NewExportHandler(db, rdb)
	importHandler := handlers.NewImportHandler(db)

	api := app.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Auth routes (protected)
	auth.Post("/logout", middleware.AuthRequired(cfg), authHandler.Logout)
	auth.Get("/me", middleware.AuthRequired(cfg), authHandler.Me)

	// Protected routes
	protected := api.Group("", middleware.AuthRequired(cfg))

	// Projects
	projects := protected.Group("/projects")
	projects.Get("/", projectHandler.List)
	projects.Post("/", projectHandler.Create)
	projects.Get("/:id", projectHandler.Get)
	projects.Put("/:id", projectHandler.Update)
	projects.Delete("/:id", projectHandler.Delete)
	projects.Get("/:id/stats", projectHandler.Stats)

	// Languages
	projects.Get("/:id/languages", languageHandler.List)
	projects.Post("/:id/languages", languageHandler.Create)
	projects.Put("/:id/languages/:langId", languageHandler.Update)
	projects.Delete("/:id/languages/:langId", languageHandler.Delete)

	// Translation keys
	projects.Get("/:id/keys", keyHandler.List)
	projects.Post("/:id/keys", keyHandler.Create)
	projects.Put("/:id/keys/:keyId", keyHandler.Update)
	projects.Delete("/:id/keys/:keyId", keyHandler.Delete)

	// Translations
	projects.Get("/:id/translations", translationHandler.Get)
	projects.Put("/:id/translations", translationHandler.BatchUpdate)

	// Import
	projects.Post("/:id/import", importHandler.Import)

	// API keys
	projects.Get("/:id/api-keys", apiKeyHandler.List)
	projects.Post("/:id/api-keys", apiKeyHandler.Create)
	projects.Delete("/:id/api-keys/:keyId", apiKeyHandler.Delete)

	// Cache management
	projects.Post("/:id/cache/invalidate", cacheHandler.Invalidate)
	projects.Get("/:id/cache/status", cacheHandler.Status)

	// Export routes (API key auth)
	export := api.Group("/export", middleware.APIKeyAuth(db))
	export.Get("/:slug/:langCode", exportHandler.Export)
}
