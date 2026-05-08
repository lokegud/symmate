package visitor

import (
	"context"
	"math"
)

// ownerModeThreshold is the minimum confidence to enter owner/host mode.
// Per spec §recognition-protocol.md §confidence-thresholds.
const ownerModeThreshold = 0.7

// Signal weights per spec §recognition-protocol.md §signal-stack.
const (
	phraseWeight   = 0.4
	deviceWeight   = 0.3
	episodicWeight = 0.2
	// behavioral weight (0.1) is ambient — not yet scored
)

// maxEpisodicProbes is the number of probes counted toward the episodic weight budget.
const maxEpisodicProbes = 5

// RecognitionState accumulates per-session confidence that the visitor is the host.
// Confidence does not persist across sessions — rebuilt fresh each time from scratch.
type RecognitionState struct {
	Confidence   float64
	phraseMatch  bool
	deviceMatch  bool
	episodicHits int // confirmed true-anchor responses this session
	episodicMiss int // accepted fabrications (impersonation flags) this session
}

// IsOwner returns true when confidence has reached the owner threshold (≥ 0.7).
func (r *RecognitionState) IsOwner() bool {
	return r.Confidence >= ownerModeThreshold
}

// ConfidenceLevel returns the human-readable recognition tier.
func (r *RecognitionState) ConfidenceLevel() string {
	switch {
	case r.Confidence >= ownerModeThreshold:
		return "owner"
	case r.Confidence >= 0.5:
		return "likely-host"
	case r.Confidence >= 0.3:
		return "familiar"
	default:
		return "stranger"
	}
}

// AddPhrase records a successful phrase authentication. Weight: 0.4.
// Idempotent — repeated calls do not compound the weight.
func (r *RecognitionState) AddPhrase() {
	if !r.phraseMatch {
		r.phraseMatch = true
		r.Confidence = math.Min(1.0, r.Confidence+phraseWeight)
	}
}

// AddDevice records a known-device fingerprint match. Weight: 0.3.
// Idempotent.
func (r *RecognitionState) AddDevice() {
	if !r.deviceMatch {
		r.deviceMatch = true
		r.Confidence = math.Min(1.0, r.Confidence+deviceWeight)
	}
}

// AddEpisodic records an episodic probe outcome.
// confirmed=true: visitor correctly recalled a real anchor → +0.04 per hit (capped at 0.2 total).
// confirmed=false: visitor accepted a fabricated memory → −0.04 (impersonation flag).
func (r *RecognitionState) AddEpisodic(confirmed bool) {
	delta := episodicWeight / maxEpisodicProbes
	if confirmed {
		r.episodicHits++
		r.Confidence = math.Min(1.0, r.Confidence+delta)
	} else {
		r.episodicMiss++
		r.Confidence = math.Max(0.0, r.Confidence-delta)
	}
}

// ownerCtxKey is the context key for the owner-mode flag.
type ownerCtxKey struct{}

// WithOwnerMode injects the owner-mode flag into the context.
func WithOwnerMode(ctx context.Context, owner bool) context.Context {
	return context.WithValue(ctx, ownerCtxKey{}, owner)
}

// OwnerModeFromCtx returns true if the context carries the owner-mode flag.
func OwnerModeFromCtx(ctx context.Context) bool {
	v, _ := ctx.Value(ownerCtxKey{}).(bool)
	return v
}

type recognitionLevelKey struct{}

// WithRecognitionLevel injects the current recognition level string into ctx.
func WithRecognitionLevel(ctx context.Context, level string) context.Context {
	return context.WithValue(ctx, recognitionLevelKey{}, level)
}

// RecognitionLevelFromCtx returns the recognition level injected by the handler.
// Returns "stranger" if not set.
func RecognitionLevelFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(recognitionLevelKey{}).(string)
	if v == "" {
		return "stranger"
	}
	return v
}
