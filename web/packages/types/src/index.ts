// Shared types for Symmate web surfaces

// Shape of GET :7842/id
export interface PersonaIdentity {
	name: string;
	slug: string;
	host_type: string;
	description: string;
	capabilities: string[];
	contact: {
		website?: string;
		location?: string;
	};
	verified: boolean;
	endpoints: {
		visitor: string;
		a2a: string;
		mcp?: string;
	};
	schema_version: string;
	updated_at: string;
}

// POST :7842/chat body
export interface ChatRequest {
	message: string;
	visitor_handle?: string;
}

// Agent Card (A2A standard) — GET :7843/.well-known/agent-card.json
export interface AgentCard {
	schemaVersion: string;
	name: string;
	description: string;
	url: string;
	capabilities: {
		streaming: boolean;
		pushNotifications: boolean;
		stateTransitionHistory: boolean;
	};
	skills: Array<{
		id: string;
		name: string;
		description: string;
		tags: string[];
	}>;
	authentication: {
		schemes: string[];
	};
}
