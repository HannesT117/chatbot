package main

import (
	"log/slog"
	"net/http"
	"os"

	"chatbot/server/internal/api"
	"chatbot/server/internal/scenario"
)

// config holds runtime configuration read from environment variables.
type config struct {
	Port         string
	OpenAIAPIKey string
}

func configFromEnv() config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return config{
		Port:         port,
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := configFromEnv()

	scenarios, err := scenario.LoadAll()
	if err != nil {
		logger.Error("failed to load scenarios", "error", err)
		os.Exit(1)
	}
	logger.Info("scenarios loaded", "count", len(scenarios))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.HandleHealth)
	mux.HandleFunc("GET /api/scenarios", api.ScenariosHandler(scenarios))

	addr := ":" + cfg.Port
	logger.Info("server starting", "addr", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
