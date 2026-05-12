package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/JMC50/nas/internal/admin"
	"github.com/JMC50/nas/internal/auth"
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

	authHandlers := &auth.Handlers{Config: cfg, DB: conn}
	adminHandlers := &admin.Handlers{Config: cfg, DB: conn}
	requireToken := auth.RequireToken(cfg.PrivateKey)

	// Public root + health
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("server is running :D"))
	})
	r.Get("/healthz", healthzHandler(conn))

	// Local auth (no token required)
	r.Post("/auth/register", authHandlers.RegisterLocal)
	r.Post("/auth/login", authHandlers.LoginLocal)
	r.Post("/auth/change-password", authHandlers.ChangePassword) // token in query, parsed inside
	r.Get("/auth/config", authHandlers.AuthConfig)

	// OAuth flows (no token required — they ISSUE tokens)
	r.Get("/login", authHandlers.DiscordLogin)
	r.Get("/kakaoLogin", authHandlers.KakaoLogin)
	r.Post("/register", authHandlers.DiscordRegister)
	r.Post("/registerKakao", authHandlers.KakaoRegister)

	// Intent inspection (legacy compat — no auth gate per legacy behavior)
	r.Get("/getIntents", authHandlers.GetIntents)
	r.Get("/getAllUsers", authHandlers.GetAllUsers)
	r.Get("/getActivityLog", adminHandlers.GetActivityLog)
	r.Post("/log", adminHandlers.InsertLog) // token in body

	// Token-required intent checks
	r.Group(func(r chi.Router) {
		r.Use(requireToken)
		r.Get("/checkAdmin", authHandlers.CheckAdmin)
		r.Get("/checkIntent", authHandlers.CheckIntent)
		r.Get("/authorize", adminHandlers.ToggleIntent)
		r.Get("/unauthorize", adminHandlers.ToggleIntent)
		r.Post("/requestAdminIntent", adminHandlers.RequestAdminIntent)
	})

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
