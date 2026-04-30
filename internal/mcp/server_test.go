package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/loke/symmate/internal/aeo"
)

func testHostContext() *aeo.HostContext {
	return &aeo.HostContext{
		Profile: &aeo.Profile{
			Name:         "TestHost",
			Slug:         "test",
			HostType:     aeo.HostTypeMoxPer,
			Description:  "A test Symmate.",
			Capabilities: []string{"chat", "profile", "catalog", "hours", "policies"},
			Endpoints: aeo.Endpoints{
				Visitor: "http://localhost:7842",
				MCP:     "http://localhost:7844",
			},
			SchemaVer: "0.1",
			UpdatedAt: time.Now(),
		},
		Catalog:  "Widget A, Widget B",
		Hours:    "Mon-Fri 9am-5pm",
		Policies: "No refunds after 30 days.",
	}
}

func sendRPC(t *testing.T, srv *Server, method string, params interface{}) rpcResponse {
	t.Helper()
	var body strings.Builder
	body.WriteString(`{"jsonrpc":"2.0","id":1,"method":"`)
	body.WriteString(method)
	body.WriteString(`"`)
	if params != nil {
		b, _ := json.Marshal(params)
		body.WriteString(`,"params":`)
		body.Write(b)
	}
	body.WriteString(`}`)

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	var resp rpcResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nbody: %s", err, w.Body.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	srv := NewServer(testHostContext(), NoopAuditLogger{})
	resp := sendRPC(t, srv, "initialize", map[string]interface{}{
		"protocolVersion": mcpProtocolVersion,
		"clientInfo":      map[string]string{"name": "test", "version": "0.1"},
	})
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result initResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if result.ProtocolVersion != mcpProtocolVersion {
		t.Errorf("protocolVersion = %q", result.ProtocolVersion)
	}
	if result.ServerInfo.Name == "" {
		t.Error("serverInfo.name should be set")
	}
}

func TestToolsList(t *testing.T) {
	srv := NewServer(testHostContext(), NoopAuditLogger{})
	resp := sendRPC(t, srv, "tools/list", nil)
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result toolsListResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	wantTools := map[string]bool{
		"get_profile":  false,
		"get_catalog":  false,
		"get_hours":    false,
		"get_policies": false,
	}
	for _, tool := range result.Tools {
		wantTools[tool.Name] = true
	}
	for name, found := range wantTools {
		if !found {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestToolCall_GetProfile(t *testing.T) {
	srv := NewServer(testHostContext(), NoopAuditLogger{})
	resp := sendRPC(t, srv, "tools/call", map[string]interface{}{
		"name":      "get_profile",
		"arguments": map[string]interface{}{},
	})
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}

	b, _ := json.Marshal(resp.Result)
	var result ToolResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	if !strings.Contains(result.Content[0].Text, "TestHost") {
		t.Errorf("profile result should contain 'TestHost', got: %s", result.Content[0].Text)
	}
}

func TestToolCall_GetCatalog(t *testing.T) {
	srv := NewServer(testHostContext(), NoopAuditLogger{})
	resp := sendRPC(t, srv, "tools/call", map[string]interface{}{
		"name": "get_catalog",
	})
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ToolResult
	json.Unmarshal(b, &result)
	if !strings.Contains(result.Content[0].Text, "Widget") {
		t.Errorf("expected catalog content, got: %s", result.Content[0].Text)
	}
}

func TestToolCall_UnknownTool(t *testing.T) {
	srv := NewServer(testHostContext(), NoopAuditLogger{})
	resp := sendRPC(t, srv, "tools/call", map[string]interface{}{
		"name": "delete_everything",
	})
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestCapabilityPolicy_Default(t *testing.T) {
	p := DefaultPolicy()
	for _, tool := range []string{"get_profile", "get_catalog", "get_hours", "get_policies"} {
		allowed, reason := p.Check(tool)
		if !allowed {
			t.Errorf("tool %q should be allowed, got denied: %s", tool, reason)
		}
	}
	allowed, _ := p.Check("not_a_tool")
	if allowed {
		t.Error("unknown tool should be denied")
	}
}

func TestPing(t *testing.T) {
	srv := NewServer(testHostContext(), NoopAuditLogger{})
	resp := sendRPC(t, srv, "ping", nil)
	if resp.Error != nil {
		t.Fatalf("ping error: %+v", resp.Error)
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := NewServer(testHostContext(), NoopAuditLogger{})
	resp := sendRPC(t, srv, "frobnicate", nil)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("error code = %d, want -32601", resp.Error.Code)
	}
}
