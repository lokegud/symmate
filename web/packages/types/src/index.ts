// Shared types for Symmate web surfaces

export interface PersonaIdentity {
	id: string;
	name: string;
	handle: string;
	description: string;
	avatarUrl?: string;
}

export interface ChatMessage {
	role: 'user' | 'assistant';
	content: string;
	timestamp: string;
}

// JSON-RPC 2.0 shape used by POST /agent/sendMessage
export interface SendMessageRequest {
	jsonrpc: '2.0';
	id: string | number;
	method: 'agent/sendMessage';
	params: {
		text: string;
		sessionId?: string;
	};
}

export interface SendMessageResponse {
	jsonrpc: '2.0';
	id: string | number;
	result?: {
		text: string;
		sessionId?: string;
	};
	error?: {
		code: number;
		message: string;
	};
}

// Agent Card (A2A standard)
export interface AgentCard {
	name: string;
	description: string;
	url: string;
	version: string;
	capabilities: {
		streaming?: boolean;
		pushNotifications?: boolean;
		stateTransitionHistory?: boolean;
	};
	skills?: Array<{
		id: string;
		name: string;
		description: string;
	}>;
}
