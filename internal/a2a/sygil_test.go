package a2a

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSenderSignal_RenderForLLM_EmptyURLNoOutput(t *testing.T) {
	if got := (SenderSignal{}).RenderForLLM(); got != "" {
		t.Errorf("empty card URL should render empty, got %q", got)
	}
}

func TestSenderSignal_RenderForLLM_Unknown(t *testing.T) {
	out := SenderSignal{
		CardURL:     "http://example.com/card.json",
		Fingerprint: "abc123",
		Known:       false,
	}.RenderForLLM()

	for _, want := range []string{
		"Sender Intel",
		"http://example.com/card.json",
		"abc123",
		"unknown to Sygil",
		"signal, not a verdict",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestSenderSignal_RenderForLLM_KnownLimited(t *testing.T) {
	until := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	out := SenderSignal{
		CardURL:      "http://example.com/card.json",
		Fingerprint:  "abc",
		Known:        true,
		Name:         "Bob",
		Description:  "test bot",
		Skills:       []string{"conversation"},
		Limited:      true,
		LimitedUntil: &until,
		FlagCount:    5,
	}.RenderForLLM()

	for _, want := range []string{
		"Status: known",
		"Bob",
		"test bot",
		"conversation",
		"Flag count (90d window): 5",
		"LIMITED",
		"2026-06-01",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestSygilClient_Lookup_NilClientReturnsUnknown(t *testing.T) {
	var c *SygilClient
	sig := c.Lookup(context.Background(), "http://example.com/card.json")
	if sig.Known {
		t.Errorf("nil client should yield unknown, got %+v", sig)
	}
	if sig.CardURL != "http://example.com/card.json" {
		t.Errorf("CardURL not preserved: %q", sig.CardURL)
	}
	if sig.Fingerprint == "" {
		t.Errorf("fingerprint should be computed even on nil client")
	}
}

func TestSygilClient_Lookup_404ReturnsUnknown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewSygilClient(srv.URL)
	sig := c.Lookup(context.Background(), "http://example.com/card.json")
	if sig.Known {
		t.Errorf("404 should yield unknown, got %+v", sig)
	}
}

func TestSygilClient_Lookup_DecodesFullPayload(t *testing.T) {
	registered := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	refreshed := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"card_url":       "http://example.com/card.json",
			"fingerprint":    "ignored-server-side",
			"registered_at":  registered,
			"last_refreshed": refreshed,
			"card": map[string]any{
				"name":        "Alice",
				"description": "test agent",
				"skill_ids":   []string{"talk", "code"},
				"tags":        []string{"general"},
			},
			"reputation": map[string]any{
				"limited": false,
				"stats":   map[string]any{"flag_count": 2},
			},
		})
	}))
	defer srv.Close()

	c := NewSygilClient(srv.URL)
	sig := c.Lookup(context.Background(), "http://example.com/card.json")
	if !sig.Known {
		t.Fatalf("expected known=true, got %+v", sig)
	}
	if sig.Name != "Alice" || sig.Description != "test agent" {
		t.Errorf("card fields not decoded: %+v", sig)
	}
	if len(sig.Skills) != 2 || sig.Skills[0] != "talk" {
		t.Errorf("skills not decoded: %+v", sig.Skills)
	}
	if sig.FlagCount != 2 {
		t.Errorf("flag_count not decoded: %d", sig.FlagCount)
	}
	if sig.Limited {
		t.Errorf("expected not limited")
	}
}
