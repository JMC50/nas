package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"time"

	"github.com/JMC50/nas/internal/admin"
	"github.com/JMC50/nas/internal/archive"
	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/files"
	"github.com/JMC50/nas/internal/stream"
	"github.com/JMC50/nas/internal/system"
	"github.com/JMC50/nas/internal/upload"
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
	fileHandlers := &files.Handlers{Config: cfg, DB: conn}
	streamHandlers := &stream.Handlers{Config: cfg, DB: conn}
	uploadHandlers := &upload.Handlers{Config: cfg, DB: conn}
	archiveTracker := archive.NewTracker(1 * time.Hour)
	archiveHandlers := &archive.Handlers{Config: cfg, DB: conn, Tracker: archiveTracker}
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
		r.Get("/stat", fileHandlers.Stat)
	})

	// File operations — token + intent
	addFileRoute(r, "GET", "/readFolder", auth.IntentView, requireToken, conn, fileHandlers.ReadFolder)
	addFileRoute(r, "GET", "/searchInAllFiles", auth.IntentView, requireToken, conn, fileHandlers.Search)
	addFileRoute(r, "GET", "/getTextFile", auth.IntentOpen, requireToken, conn, fileHandlers.GetTextFile)
	addFileRoute(r, "POST", "/saveTextFile", auth.IntentUpload, requireToken, conn, fileHandlers.SaveTextFile)
	addFileRoute(r, "GET", "/makedir", auth.IntentUpload, requireToken, conn, fileHandlers.MakeDir)
	addFileRoute(r, "GET", "/forceDelete", auth.IntentDelete, requireToken, conn, fileHandlers.ForceDelete)
	addFileRoute(r, "GET", "/copy", auth.IntentCopy, requireToken, conn, fileHandlers.Copy)
	addFileRoute(r, "GET", "/move", auth.IntentCopy, requireToken, conn, fileHandlers.Move)
	addFileRoute(r, "GET", "/rename", auth.IntentRename, requireToken, conn, fileHandlers.Rename)

	// Streaming + download
	addFileRoute(r, "GET", "/getVideoData", auth.IntentOpen, requireToken, conn, streamHandlers.Video)
	addFileRoute(r, "GET", "/getAudioData", auth.IntentOpen, requireToken, conn, streamHandlers.Audio)
	addFileRoute(r, "GET", "/getImageData", auth.IntentOpen, requireToken, conn, streamHandlers.Image)
	addFileRoute(r, "GET", "/download", auth.IntentDownload, requireToken, conn, streamHandlers.Download)
	r.Get("/img", streamHandlers.Img) // bundled icons — no auth (matches legacy)

	// Legacy upload wrappers (raw body stream)
	addFileRoute(r, "POST", "/input", auth.IntentUpload, requireToken, conn, uploadHandlers.LegacyInput)
	addFileRoute(r, "POST", "/inputZip", auth.IntentUpload, requireToken, conn, uploadHandlers.LegacyInputZip)

	// Archive operations
	addFileRoute(r, "POST", "/zipFiles", auth.IntentUpload, requireToken, conn, archiveHandlers.ZipFiles)
	addFileRoute(r, "POST", "/unzipFile", auth.IntentUpload, requireToken, conn, archiveHandlers.UnzipFile)
	r.Get("/progress", archiveHandlers.Progress)
	r.Group(func(r chi.Router) {
		r.Use(requireToken)
		r.Get("/downloadZip", archiveHandlers.DownloadZip)
		r.Get("/deleteTempZip", archiveHandlers.DeleteTempZip)
	})

	// System info (no auth — legacy compat)
	r.Get("/getSystemInfo", system.GetSystemInfoHandler)

	// tus resumable upload protocol at /files/*
	stagingDir := filepath.Join(cfg.NASTempDir, "tus")
	tusHandler, err := uploadHandlers.MountTus(stagingDir)
	if err == nil {
		r.Handle("/files/*", http.StripPrefix("/files", tusHandler))
		r.Handle("/files", http.StripPrefix("/files", tusHandler))
	}

	return r
}

func addFileRoute(
	r chi.Router,
	method, path string,
	intent auth.Intent,
	requireToken func(http.Handler) http.Handler,
	conn *sql.DB,
	handler http.HandlerFunc,
) {
	chain := requireToken(auth.RequireIntent(conn, intent)(handler))
	r.Method(method, path, chain)
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
