package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
)

func NewRouter(cfg *config.Config, conn *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CorsOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Tus-Resumable", "Upload-Length", "Upload-Metadata", "Upload-Offset"},
		ExposedHeaders:   []string{"Tus-Resumable", "Upload-Offset", "Upload-Length", "Location"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("server is running :D"))
	})

	r.Get("/healthz", healthzHandler(conn))
	return r
}

func healthzHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{"status": "ok"}
		if err := conn.Ping(); err != nil {
			response["status"] = "degraded"
			response["db"] = "disconnected"
		} else {
			response["db"] = "connected"
		}
		if err := db.VerifySchema(conn); err != nil {
			response["schema"] = "invalid: " + err.Error()
		} else {
			response["schema"] = "valid"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
