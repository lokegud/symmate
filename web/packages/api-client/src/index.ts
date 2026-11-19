import type { AgentCard, PersonaIdentity } from '@symmate/types';

// Owner API base — :7841, loopback only (called from SvelteKit server-side).
export const OWNER_PORT = 7841;
export function ownerBase(host = '127.0.0.1'): string {
	return `http://${host}:${OWNER_PORT}`;
}

// --- Owner API client ---

export interface OwnerSummary {
	name: string;
	enrolled: boolean;
	conversation_count: number;
	devices: number;
}

export interface PersonalityData {
	context: string;
	vertical: string;
}

export interface PolicyData {
	tool_tags: string[];
}

export interface LLMConfig {
	provider: string;
	model: string;
	custom_endpoint: string;
	has_custom_key: boolean;
}

export interface AuditPage {
	entries: Record<string, unknown>[];
	page: number;
	total: number;
	pages: number;
}

export interface ConversationSummary {
	id: string;
	visitor_handle?: string;
	surface: string;
	started_at: number;
	ended_at?: number;
	outcome?: string;
}

export interface ConversationDetail extends ConversationSummary {
	messages: { role: string; content: string; ts: number }[];
}

function ownerFetch(token: string, path: string, opts?: RequestInit): Promise<Response> {
	return fetch(`${ownerBase()}${path}`, {
		...opts,
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`,
			...(opts?.headers ?? {})
		}
	});
}

export async function ownerLogin(phrase: string): Promise<{ token: string }> {
	const res = await fetch(`${ownerBase()}/auth/login`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ phrase })
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({}));
		throw new Error((err as { error?: string }).error ?? `login: ${res.status}`);
	}
	return res.json();
}

export async function ownerLogout(token: string): Promise<void> {
	await fetch(`${ownerBase()}/auth/logout`, {
		method: 'POST',
		headers: { Authorization: `Bearer ${token}` }
	});
}

export async function ownerSummary(token: string): Promise<OwnerSummary> {
	const res = await ownerFetch(token, '/summary');
	if (!res.ok) throw new Error(`summary: ${res.status}`);
	return res.json();
}

export async function ownerGetPersonality(token: string): Promise<PersonalityData> {
	const res = await ownerFetch(token, '/personality');
	if (!res.ok) throw new Error(`getPersonality: ${res.status}`);
	return res.json();
}

export async function ownerPutPersonality(token: string, data: Partial<PersonalityData>): Promise<void> {
	const res = await ownerFetch(token, '/personality', {
		method: 'PUT',
		body: JSON.stringify(data)
	});
	if (!res.ok) throw new Error(`putPersonality: ${res.status}`);
}

export async function ownerGetPolicy(token: string): Promise<PolicyData> {
	const res = await ownerFetch(token, '/policy');
	if (!res.ok) throw new Error(`getPolicy: ${res.status}`);
	return res.json();
}

export async function ownerPutPolicy(token: string, data: PolicyData): Promise<void> {
	const res = await ownerFetch(token, '/policy', {
		method: 'PUT',
		body: JSON.stringify(data)
	});
	if (!res.ok) throw new Error(`putPolicy: ${res.status}`);
}

export async function ownerGetAudit(token: string, page = 1): Promise<AuditPage> {
	const res = await ownerFetch(token, `/audit?page=${page}`);
	if (!res.ok) throw new Error(`getAudit: ${res.status}`);
	return res.json();
}

export async function ownerListConversations(token: string, page = 1): Promise<{ conversations: ConversationSummary[]; page: number }> {
	const res = await ownerFetch(token, `/conversations?page=${page}`);
	if (!res.ok) throw new Error(`listConversations: ${res.status}`);
	return res.json();
}

export async function ownerGetConversation(token: string, id: string): Promise<ConversationDetail> {
	const res = await ownerFetch(token, `/conversations/${id}`);
	if (!res.ok) throw new Error(`getConversation: ${res.status}`);
	return res.json();
}

export async function ownerGetLLM(token: string): Promise<LLMConfig> {
	const res = await ownerFetch(token, '/llm');
	if (!res.ok) throw new Error(`getLLM: ${res.status}`);
	return res.json();
}

export async function ownerPutLLM(token: string, cfg: Omit<LLMConfig, 'has_custom_key'> & { custom_api_key?: string }): Promise<void> {
	const res = await ownerFetch(token, '/llm', {
		method: 'PUT',
		body: JSON.stringify(cfg)
	});
	if (!res.ok) throw new Error(`putLLM: ${res.status}`);
}

// --- Registered users (per-user tool access) ---

export interface RegisteredUser {
	id: string;
	name: string;
	enabled: boolean;
	groups: string[];
	created_at: string;
	last_seen?: string;
}

export interface UsersData {
	/** Master switch: when false every non-owner is an anonymous visitor. */
	enabled: boolean;
	/** Restricted tool groups that can be granted (e.g. "engineering"). */
	available_groups: string[];
	users: RegisteredUser[];
}

async function ownerJSON<T>(res: Response, label: string): Promise<T> {
	if (!res.ok) {
		let msg = `${label}: ${res.status}`;
		try {
			const body = await res.json();
			if (body?.error) msg = `${label}: ${body.error}`;
		} catch {
			/* non-JSON error body */
		}
		throw new Error(msg);
	}
	return res.status === 204 ? (undefined as T) : res.json();
}

export async function ownerListUsers(token: string): Promise<UsersData> {
	return ownerJSON(await ownerFetch(token, '/users'), 'listUsers');
}

export async function ownerSetUsersEnabled(token: string, enabled: boolean): Promise<void> {
	return ownerJSON(
		await ownerFetch(token, '/users/settings', { method: 'PUT', body: JSON.stringify({ enabled }) }),
		'setUsersEnabled'
	);
}

/** Returns the new user and its access token. The token is only ever shown here. */
export async function ownerCreateUser(token: string, name: string, groups: string[]): Promise<{ user: RegisteredUser; token: string }> {
	return ownerJSON(
		await ownerFetch(token, '/users', { method: 'POST', body: JSON.stringify({ name, groups }) }),
		'createUser'
	);
}

export async function ownerUpdateUser(
	token: string,
	id: string,
	patch: { name?: string; enabled?: boolean; groups?: string[] }
): Promise<RegisteredUser> {
	return ownerJSON(
		await ownerFetch(token, `/users/${encodeURIComponent(id)}`, { method: 'PATCH', body: JSON.stringify(patch) }),
		'updateUser'
	);
}

export async function ownerRotateUserToken(token: string, id: string): Promise<{ token: string }> {
	return ownerJSON(
		await ownerFetch(token, `/users/${encodeURIComponent(id)}/token`, { method: 'POST' }),
		'rotateUserToken'
	);
}

export async function ownerDeleteUser(token: string, id: string): Promise<void> {
	return ownerJSON(
		await ownerFetch(token, `/users/${encodeURIComponent(id)}`, { method: 'DELETE' }),
		'deleteUser'
	);
}

// Runtime ports — overridden by SYMMATE_VISITOR_PORT env var in server context
const _visitorPort = (typeof process !== 'undefined' && process.env?.SYMMATE_VISITOR_PORT)
	? parseInt(process.env.SYMMATE_VISITOR_PORT, 10)
	: 7842;
export const VISITOR_PORT = _visitorPort;
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
