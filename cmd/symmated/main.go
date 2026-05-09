package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/loke/symmate/internal/a2a"
	"github.com/loke/symmate/internal/aeo"
	"github.com/loke/symmate/internal/audit"
	"github.com/loke/symmate/internal/config"
	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/llm"
	"github.com/loke/symmate/internal/logger"
	"github.com/loke/symmate/internal/mcp"
	"github.com/loke/symmate/internal/owner"
	"github.com/loke/symmate/internal/personality"
	"github.com/loke/symmate/internal/storage"
	"github.com/loke/symmate/internal/tool"
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

	// --- Storage (acquires singleton lock, runs migrations) ---
	key, err := loadOrDeriveKey(cfg.Host.HostDir)
	if err != nil {
		l.Fatal().Err(err).Msg("key setup failed")
	}
	db, err := storage.Open(cfg.Host.HostDir, key)
	if err != nil {
		l.Fatal().Err(err).Msg("storage open failed — is another symmated running for this host?")
	}
	defer db.Close()

	// --- Audit log (JSONL primary + SQLite encrypted mirror) ---
	auditPath := filepath.Join(cfg.Host.HostDir, "audit.jsonl")
	auditStore, err := audit.NewSQLiteStore(auditPath, &auditStoreBridge{db: db})
	if err != nil {
		l.Fatal().Err(err).Msg("audit store init failed")
	}
	auditLogger := audit.NewLogger(auditStore)
	defer auditLogger.Close()

	// --- Personality engine (hot-reload on context.md / vertical.md changes) ---
	personalityLoader, err := personality.NewLoader(cfg.Host.HostDir)
	if err != nil {
		l.Fatal().Err(err).Str("hostDir", cfg.Host.HostDir).Msg("personality loader init failed")
	}
	defer personalityLoader.Close()

	// --- LLM override config (owner-configurable) ---
	llmStore := llm.NewOverrideStore(cfg.Host.HostDir)

	// --- LLM router ---
	router := llm.NewRouter(&llmAuditBridge{al: auditLogger})
	router.WithOverride(llmStore)

	// --- Privacy firewall ---
	fw := firewall.New(&firewallAuditBridge{al: auditLogger})

	// --- Tool registry ---
	toolReg := &tool.Registry{}
	toolReg.Register(tool.DateTimeTool())
	toolReg.Register(tool.FetchURLTool())
	if searchURL := os.Getenv("SYMMATE_SEARCH_URL"); searchURL != "" {
		toolReg.Register(tool.SearchTool(searchURL))
		l.Info().Str("search_url", searchURL).Msg("web_search tool enabled")
	}
	l.Info().Int("tools", toolReg.Len()).Msg("tool registry ready")

	// --- Owner config (ToolTags allowlist) ---
	ownerCfg, err := visitor.LoadOwnerConfig(cfg.Host.HostDir)
	if err != nil {
		l.Warn().Err(err).Msg("could not load owner config; all tools allowed")
		ownerCfg = &visitor.OwnerConfig{}
	}

	// --- AEO profile ---
	// Per-surface public URL overrides (HTTPS in production per A2A spec §7.1).
	// Falls back to BaseURL:Port when no override set — typically plain HTTP, dev/internal only.
	visitorBase := cfg.Host.VisitorPublicURL
	if visitorBase == "" {
		visitorBase = fmt.Sprintf("%s:%d", cfg.Host.BaseURL, cfg.Visitor.Port)
	}
	a2aBase := cfg.Host.A2APublicURL
	if a2aBase == "" {
		a2aBase = fmt.Sprintf("%s:%d", cfg.Host.BaseURL, cfg.A2A.Port)
	}
	mcpBase := cfg.Host.MCPPublicURL
	if mcpBase == "" {
		mcpBase = fmt.Sprintf("%s:%d", cfg.Host.BaseURL, cfg.MCP.Port)
	}
	hc, err := aeo.Build(aeo.BuildConfig{
		Slug:        cfg.Host.Slug,
		HostDir:     cfg.Host.HostDir,
		HostType:    aeo.HostTypeMoxPer,
		VisitorBase: visitorBase,
		A2ABase:     a2aBase,
		MCPBase:     mcpBase,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("failed to build host context")
	}
	l.Info().Str("name", hc.Profile.Name).Msg("host context loaded")

	// --- :8080 Legacy health/version server ---
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", handleHealth)
	healthMux.HandleFunc("/version", handleVersion)
	healthServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: healthMux,
	}

	// --- Test-trace logger (tailed by Loke during chat tests) ---
	traceLog, err := visitor.NewTraceLogger(filepath.Join(cfg.Host.HostDir, "test-trace.jsonl"))
	if err != nil {
		l.Warn().Err(err).Msg("trace logger disabled — could not open test-trace.jsonl")
	} else {
		defer traceLog.Close()
		l.Info().Str("path", filepath.Join(cfg.Host.HostDir, "test-trace.jsonl")).Msg("test-trace logging active")
	}

	// --- :7842 Visitor API ---
	brain := &Brain{
		hostDir:  cfg.Host.HostDir,
		loader:   personalityLoader,
		router:   router,
		firewall: fw,
		grants:   []firewall.Grant{},
		tools:    toolReg,
		toolTags: ownerCfg.ToolTags,
		trace:    traceLog,
	}
	sessions := visitor.NewSessions(cfg.Visitor.RatePerMin, cfg.Visitor.TokenCap)
	var ff *visitor.Forcefield
	if cfg.Visitor.PokerEndpoint != "" {
		ff = visitor.NewForcefield(cfg.Visitor.PokerEndpoint)
	}
	visitorServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Visitor.Port),
		Handler: visitor.NewRouter(hc, brain, sessions, ff, storage.NewVisitorStore(db), visitor.NewOwnerManager(cfg.Host.HostDir)),
	}

	// --- :7841 Owner API ---
	ownerServer := owner.New(cfg.Host.HostDir, visitor.NewOwnerManager(cfg.Host.HostDir), db, llmStore)
	ownerAPIServer := &http.Server{
		Addr:    "127.0.0.1:7841",
		Handler: ownerServer.Router(),
	}

	// --- A2A endpoint (default + optional Yggdrasil extra listener) ---
	a2aCard := a2a.BuildCard(hc.Profile.Endpoints.A2A, hc.Profile.Name, hc.Profile.Description)
	a2aExecutor := a2a.NewSymmateExecutor(personalityLoader, router, fw)
	if sygilURL := os.Getenv("SYGIL_URL"); sygilURL != "" {
		a2aExecutor.WithSygil(a2a.NewSygilClient(sygilURL))
		l.Info().Str("sygil_url", sygilURL).Msg("sygil reputation check enabled")
	}
	a2aHandler := a2a.NewHandler(a2aCard, a2aExecutor)
	a2aServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.A2A.Port),
		Handler: a2aHandler,
	}

	// --- :7844 MCP server ---
	mcpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.MCP.Port),
		Handler: mcp.NewServer(hc, &mcpAuditBridge{al: auditLogger}).Router(),
	}

	servers := []*http.Server{healthServer, visitorServer, ownerAPIServer, a2aServer, mcpServer}
	if cfg.A2A.YggAddr != "" {
		l.Info().Str("ygg_addr", cfg.A2A.YggAddr).
			Str("a2a_port", fmt.Sprintf("%d", cfg.A2A.Port)).
			Msg("a2a reachable via yggdrasil (dual-stack listener)")
	}
	for _, srv := range servers {
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
	for _, srv := range servers {
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
