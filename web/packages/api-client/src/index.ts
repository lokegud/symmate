import type { AgentCard, PersonaIdentity } from '@symmate/types';

// Runtime ports — can be overridden by SYMMATE_VISITOR_PORT env var in server context
export const VISITOR_PORT = 7842;
export const AGENT_CARD_PORT = 7843;

export function visitorBase(host = '127.0.0.1'): string {
	return `http://${host}:${VISITOR_PORT}`;
}

export async function getPersonaIdentity(host?: string): Promise<PersonaIdentity> {
	const res = await fetch(`${visitorBase(host)}/id`);
	if (!res.ok) throw new Error(`getPersonaIdentity: ${res.status}`);
	return res.json();
}

export async function getAgentCard(host?: string): Promise<AgentCard> {
	const res = await fetch(`${visitorBase(host)}/.well-known/agent-card.json`);
	if (!res.ok) throw new Error(`getAgentCard: ${res.status}`);
	return res.json();
}

// Returns a raw Response with SSE body — caller pipes it to the client.
export async function startChatStream(message: string, host?: string): Promise<Response> {
	return fetch(`${visitorBase(host)}/chat`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ message })
	});
}
