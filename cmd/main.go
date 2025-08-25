package main

import (
	"log/slog"
	"os"
	"wb-examples-l0/internal/config"
	"wb-examples-l0/internal/lib/logger/sl"
)

func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("config", cfg)

	log.Info("starting server url-shortener",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.HTTPServer.Address),
	)

}
