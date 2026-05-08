package visitor

import "context"

const ownerScoreThreshold = 50

// RecognitionState accumulates confidence that the current session is the owner.
// Signals add to a score; IsOwner fires above the threshold.
type RecognitionState struct {
	Score       int
	PhraseMatch bool
	DeviceMatch bool
}

// IsOwner returns true when confidence exceeds the threshold.
func (r *RecognitionState) IsOwner() bool {
	return r.Score >= ownerScoreThreshold
}

// AddPhrase records a correct phrase match (+50 confidence).
func (r *RecognitionState) AddPhrase() {
	if !r.PhraseMatch {
		r.PhraseMatch = true
		r.Score += 50
	}
}

// AddDevice records a known-device match (+35 confidence).
func (r *RecognitionState) AddDevice() {
	if !r.DeviceMatch {
		r.DeviceMatch = true
		r.Score += 35
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
