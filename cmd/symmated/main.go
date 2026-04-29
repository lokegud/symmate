package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/loke/symmate/internal/aeo"
	"github.com/loke/symmate/internal/config"
	"github.com/loke/symmate/internal/logger"
	"github.com/loke/symmate/internal/mcp"
	"github.com/loke/symmate/internal/visitor"
	"github.com/loke/symmate/pkg/version"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Setup()
	l := log.With().Str("component", "main").Logger()

	cfg, err := config.Load()
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load config")
	}

	l.Info().Str("version", version.String()).Str("host", cfg.Host.Slug).Msg("starting symmated")

	// Build AEO profile + host context from context.md.
	hc, err := aeo.Build(aeo.BuildConfig{
		Slug:        cfg.Host.Slug,
		HostDir:     cfg.Host.HostDir,
		HostType:    aeo.HostTypeMoxPer,
		VisitorBase: fmt.Sprintf("%s:%d", cfg.Host.BaseURL, cfg.Visitor.Port),
		A2ABase:     fmt.Sprintf("%s:%d", cfg.Host.BaseURL, 7843),
		MCPBase:     fmt.Sprintf("%s:%d", cfg.Host.BaseURL, cfg.MCP.Port),
	})
	if err != nil {
		l.Fatal().Err(err).Msg("failed to build host context")
	}
	l.Info().Str("name", hc.Profile.Name).Msg("host context loaded")

	// Legacy health/version server.
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", handleHealth)
	healthMux.HandleFunc("/version", handleVersion)
	healthServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: healthMux,
	}

	// :7842 Visitor API.
	brain := visitor.BrainStub{} // TODO: replace with Forge's brain when feature/brain merges
	sessions := visitor.NewSessions(cfg.Visitor.RatePerMin, cfg.Visitor.TokenCap)
	visitorServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Visitor.Port),
		Handler: visitor.NewRouter(hc, brain, sessions),
	}

	// :7844 MCP server.
	mcpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.MCP.Port),
		Handler: mcp.NewServer(hc, nil).Router(),
	}

	for _, srv := range []*http.Server{healthServer, visitorServer, mcpServer} {
		s := srv
		go func() {
			l.Info().Str("addr", s.Addr).Msg("listening")
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				l.Fatal().Err(err).Str("addr", s.Addr).Msg("server error")
			}
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	l.Info().Msg("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, srv := range []*http.Server{healthServer, visitorServer, mcpServer} {
		if err := srv.Shutdown(ctx); err != nil {
			l.Error().Err(err).Str("addr", srv.Addr).Msg("shutdown error")
		}
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok\n")
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "version: %s\n", version.String())
}
