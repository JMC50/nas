package main

import (
	"log/slog"
	"os"

	"github.com/JMC50/nas/internal/config"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/server"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("config error", "err", err)
		os.Exit(1)
	}
	if err := cfg.ResolvePaths(); err != nil {
		slog.Error("path resolution failed", "err", err)
		os.Exit(1)
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := db.VerifySchema(conn); err != nil {
		slog.Error("schema verification failed", "err", err)
		os.Exit(1)
	}

	if err := server.Run(cfg, conn); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
