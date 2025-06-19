package transport

import (
	"log/slog"
	"memorydb/api"
	"memorydb/internal/db"
	"memorydb/internal/transport/schemas"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

// mountRouter mounts the main router with all sub-routers and middlewares.
func mountRouter(logger *slog.Logger, db db.DBClient) http.Handler {
	r := chi.NewRouter()

	// add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// mount v1 router
	r.Mount("/api/v1", mountRouterV1(logger, db))

	return r
}

// mountRouterV1 mounts the v1 router with its specific routes. In this project, there are not going to be more versions,
// but this approach shows how we could handle versioning in other projects.
func mountRouterV1(logger *slog.Logger, db db.DBClient) http.Handler {
	r := chi.NewRouter()

	// start handler
	h := NewHandler(logger, db)

	r.Post("/set", h.HandleSet)
	r.Get("/{key}", h.HandleGet)
	r.Delete("/{key}", h.HandleRemove)
	r.Patch("/{key}", h.HandleUpdate)
	r.Patch("/{key}/push", h.HandlePush)
	r.Patch("/{key}/pop", h.HandlePop)

	// serve swagger UI

	r.Get("/docs/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, api.SwaggerSpec, "swagger_v1.yaml")
	})

	r.Get("/docs/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/api/v1/docs/swagger.yaml"), // The url pointing to API definition
	))

	return r
}

// mountHealthRouter mounts the health check router.
func mountHealthRouter(logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Health check endpoint hit")
		writeJSON(w, http.StatusOK, schemas.OKResponse{Message: "ok"})
	})

	return r
}
