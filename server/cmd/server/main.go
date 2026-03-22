package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"chatbot/server/internal/api"
	"chatbot/server/internal/llm"
	"chatbot/server/internal/scenario"
	"chatbot/server/internal/session"
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

	if cfg.OpenAIAPIKey == "" {
		logger.Warn("OPENAI_API_KEY is not set; LLM calls will fail")
	}

	scenarios, err := scenario.LoadAll()
	if err != nil {
		logger.Error("failed to load scenarios", "error", err)
		os.Exit(1)
	}
	logger.Info("scenarios loaded", "count", len(scenarios))

	store := session.NewInMemoryStore()
	const model = "gpt-4o-mini"
	llmClient := llm.NewOpenAIClient(cfg.OpenAIAPIKey, "", model)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.HandleHealth)
	mux.HandleFunc("GET /api/scenarios", api.ScenariosHandler(scenarios))
	mux.HandleFunc("POST /api/sessions", api.CreateSessionHandler(store, scenarios, logger))
	mux.HandleFunc("DELETE /api/sessions/{id}", api.DeleteSessionHandler(store, logger))
	mux.HandleFunc("POST /api/chat", api.ChatHandler(store, scenarios, llmClient, model, logger))

	addr := ":" + cfg.Port
	logger.Info("server starting", "addr", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      api.CORSMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second, // long enough for LLM streaming
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
