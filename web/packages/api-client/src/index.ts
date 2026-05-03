import type { AgentCard, PersonaIdentity, SendMessageRequest, SendMessageResponse } from '@symmate/types';

// Port constants — override via VITE_* env vars in each app
const VISITOR_PORT = 7842;
const OWNER_PORT = 7841;
const AGENT_CARD_PORT = 7843;

function base(port: number): string {
	const host = typeof window !== 'undefined' ? window.location.hostname : 'localhost';
	return `http://${host}:${port}`;
}

let _rpcId = 0;

// Visitor API (:7842)

export async function getPersonaIdentity(): Promise<PersonaIdentity> {
	const res = await fetch(`${base(VISITOR_PORT)}/id`);
	if (!res.ok) throw new Error(`getPersonaIdentity: ${res.status}`);
	return res.json();
}

export async function sendMessage(
	text: string,
	sessionId?: string
): Promise<SendMessageResponse> {
	const body: SendMessageRequest = {
		jsonrpc: '2.0',
		id: ++_rpcId,
		method: 'agent/sendMessage',
		params: { text, sessionId }
	};
	const res = await fetch(`${base(VISITOR_PORT)}/agent/sendMessage`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (!res.ok) throw new Error(`sendMessage: ${res.status}`);
	return res.json();
}

// Agent Card API (:7843)

export async function getAgentCard(): Promise<AgentCard> {
	const res = await fetch(`${base(AGENT_CARD_PORT)}/.well-known/agent-card.json`);
	if (!res.ok) throw new Error(`getAgentCard: ${res.status}`);
	return res.json();
}

export { VISITOR_PORT, OWNER_PORT, AGENT_CARD_PORT };
