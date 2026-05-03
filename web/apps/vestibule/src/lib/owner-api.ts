// Owner API client — server-side only (localhost:7841).
// Never call this from browser-side code; import only in +page.server.ts / +server.ts files.

const OWNER_PORT = process.env.SYMMATE_OWNER_PORT ?? '7841';
const BASE = `http://127.0.0.1:${OWNER_PORT}`;

async function ownerFetch(path: string, init?: RequestInit): Promise<Response> {
	const res = await fetch(`${BASE}${path}`, init);
	if (!res.ok) {
		const text = await res.text().catch(() => '');
		throw new Error(`owner API ${path}: ${res.status} ${text}`);
	}
	return res;
}

// --- personality ---

export type PersonalityFiles = Record<string, string>;

export async function getPersonality(): Promise<PersonalityFiles> {
	const res = await ownerFetch('/personality');
	return res.json();
}

export async function putPersonalityFile(name: string, content: string): Promise<void> {
	await ownerFetch(`/personality/${name}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ content })
	});
}

// --- grants ---

export interface Grant {
	id: number;
	scope: string;
	data_class: string;
	permission: 'allow' | 'deny' | 'ask';
	granted_at: number;
	granted_by: string;
	expires_at: number | null;
}

export async function listGrants(scope?: string): Promise<Grant[]> {
	const qs = scope ? `?scope=${encodeURIComponent(scope)}` : '';
	const res = await ownerFetch(`/policy/grants${qs}`);
	return res.json();
}

export async function createGrant(
	scope: string,
	data_class: string,
	permission: 'allow' | 'deny' | 'ask',
	expires_at?: number
): Promise<Grant> {
	const res = await ownerFetch('/policy/grants', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ scope, data_class, permission, expires_at })
	});
	return res.json();
}

export async function deleteGrant(id: number): Promise<void> {
	await ownerFetch(`/policy/grants/${id}`, { method: 'DELETE' });
}

// --- audit ---

export interface AuditEvent {
	id: string;
	timestamp: string;
	category: string;
	severity: 'info' | 'notice' | 'warning' | 'alert';
	conversation_id?: string;
	payload: unknown;
}

export async function listAuditEvents(params?: {
	severity?: string;
	category?: string;
	since?: string;
	limit?: number;
}): Promise<AuditEvent[]> {
	const qs = new URLSearchParams();
	if (params?.severity) qs.set('severity', params.severity);
	if (params?.category) qs.set('category', params.category);
	if (params?.since) qs.set('since', params.since);
	if (params?.limit) qs.set('limit', String(params.limit));
	const res = await ownerFetch(`/audit/events${qs.size ? '?' + qs : ''}`);
	return res.json();
}

// --- conversations ---

export interface Conversation {
	id: string;
	visitor_handle: string | null;
	surface: string;
	started_at: number;
	ended_at: number | null;
	outcome: string | null;
}

export interface Message {
	id: number;
	conversation_id: string;
	role: string;
	content: string;
	ts: number;
	llm_provider: string | null;
	tokens_in: number | null;
	tokens_out: number | null;
	flags: string | null;
}

export async function listConversations(limit = 50): Promise<Conversation[]> {
	const res = await ownerFetch(`/conversations?limit=${limit}`);
	return res.json();
}

export async function getConversation(id: string): Promise<{ conversation: Conversation; messages: Message[] }> {
	const res = await ownerFetch(`/conversations/${id}`);
	return res.json();
}

export async function deleteConversation(id: string): Promise<void> {
	await ownerFetch(`/conversations/${id}`, { method: 'DELETE' });
}

// --- control ---

export async function pauseSymmate(): Promise<{ paused: boolean }> {
	const res = await ownerFetch('/control/pause', { method: 'POST' });
	return res.json();
}

export async function unpauseSymmate(): Promise<{ paused: boolean }> {
	const res = await ownerFetch('/control/unpause', { method: 'POST' });
	return res.json();
}

export async function getStatus(): Promise<{ paused: boolean }> {
	const res = await ownerFetch('/control/status');
	return res.json();
}
