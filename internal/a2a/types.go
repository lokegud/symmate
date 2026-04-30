package a2a

// JSON-RPC 2.0 types.

// Request is an inbound JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id"`
}

// Response is an outbound JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standard JSON-RPC error codes.
const (
	ErrParse          = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
	ErrInjection      = -32000 // Symmate extension: injection attempt detected
)

// AgentCard is the A2A Agent Card served at /.well-known/agent-card.json.
// Structure follows A2A v1.0 protocol.
type AgentCard struct {
	SchemaVersion  string          `json:"schemaVersion"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	URL            string          `json:"url"`
	Capabilities   CardCapabilities `json:"capabilities"`
	Skills         []CardSkill      `json:"skills"`
	Authentication CardAuth         `json:"authentication"`
}

// CardCapabilities describes what the agent supports.
type CardCapabilities struct {
	Streaming              bool `json:"streaming"`
	PushNotifications      bool `json:"pushNotifications"`
	StateTransitionHistory bool `json:"stateTransitionHistory"`
}

// CardSkill is a discrete capability the agent advertises.
type CardSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
}

// CardAuth describes authentication schemes the agent accepts.
type CardAuth struct {
	Schemes []string `json:"schemes"`
}

// SendMessageParams is the params object for agent/sendMessage.
type SendMessageParams struct {
	From     string                 `json:"from,omitempty"`
	Intent   string                 `json:"intent,omitempty"`
	Text     string                 `json:"text,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Optional future field for three-party signing (feature/a2a-signing).
	Signatures []interface{} `json:"signatures,omitempty"`
}

// SendMessageResult is the result object for agent/sendMessage.
type SendMessageResult struct {
	Text     string `json:"text"`
	Filtered bool   `json:"filtered,omitempty"` // true if privacy firewall redacted content
}
