package llm_test

import (
	"context"
	"testing"
	"time"

	"github.com/loke/symmate/internal/llm"
	"github.com/loke/symmate/internal/llm/stub"
)

func TestAnthropicProvider_OK(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetReply("here is my answer")

	p := llm.NewAnthropicProviderAt(srv.URL+"/v1/messages", "test-key")
	got, err := p.Complete(context.Background(), llm.ModelSonnet, []llm.Message{
		{Role: "system", Content: "You are a test assistant."},
		{Role: "user", Content: "Hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "here is my answer" {
		t.Errorf("got %q, want %q", got, "here is my answer")
	}
}

func TestAnthropicProvider_ErrorResponse(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetError(401, "authentication_error: x-api-key header is required")

	p := llm.NewAnthropicProviderAt(srv.URL+"/v1/messages", "bad-key")
	_, err := p.Complete(context.Background(), llm.ModelSonnet, []llm.Message{
		{Role: "user", Content: "Hello"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAnthropicProvider_RateLimitError(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetError(429, "rate_limit_error: too many requests")

	p := llm.NewAnthropicProviderAt(srv.URL+"/v1/messages", "key")
	_, err := p.Complete(context.Background(), llm.ModelSonnet, []llm.Message{
		{Role: "user", Content: "Hello"},
	})
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}
}

func TestAnthropicProvider_ErrorThenRecover(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetReply("success")
	srv.SetError(500, "server error")

	p := llm.NewAnthropicProviderAt(srv.URL+"/v1/messages", "key")
	msgs := []llm.Message{{Role: "user", Content: "hi"}}

	// First call hits the error
	_, err := p.Complete(context.Background(), llm.ModelSonnet, msgs)
	if err == nil {
		t.Fatal("expected error on first call")
	}
	// Second call should succeed (error was consumed)
	got, err := p.Complete(context.Background(), llm.ModelSonnet, msgs)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if got != "success" {
		t.Errorf("got %q, want %q", got, "success")
	}
}

func TestHermesProvider_OK(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetReply("hermes says hello")

	p := llm.NewHermesProvider(srv.URL+"/v1", "stak-key")
	got, err := p.Complete(context.Background(), "hermes", []llm.Message{
		{Role: "user", Content: "Hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hermes says hello" {
		t.Errorf("got %q, want %q", got, "hermes says hello")
	}
}

func TestHermesProvider_ErrorResponse(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetError(401, "invalid api key")

	p := llm.NewHermesProvider(srv.URL+"/v1", "bad-key")
	_, err := p.Complete(context.Background(), "hermes", []llm.Message{
		{Role: "user", Content: "Hello"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStubLatency(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetReply("slow")
	srv.SetLatency(50 * time.Millisecond)

	p := llm.NewHermesProvider(srv.URL+"/v1", "key")
	start := time.Now()
	_, err := p.Complete(context.Background(), "hermes", []llm.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) < 50*time.Millisecond {
		t.Error("response faster than injected latency")
	}
}

func TestStubContextCancel(t *testing.T) {
	srv := stub.New()
	defer srv.Close()
	srv.SetLatency(500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	p := llm.NewAnthropicProviderAt(srv.URL+"/v1/messages", "key")
	_, err := p.Complete(ctx, llm.ModelSonnet, []llm.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected context timeout error, got nil")
	}
}
