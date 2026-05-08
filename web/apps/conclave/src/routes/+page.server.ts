import { env } from '$env/dynamic/private';
import type { PageServerLoad } from './$types';

const SYGIL_URL = env.SYGIL_URL ?? 'http://127.0.0.1:7448';

export interface SigilAgent {
	fingerprint: string;
	card_url: string;
	card: {
		name: string;
		description?: string;
		version?: string;
		skill_ids?: string[];
		tags?: string[];
	} | null;
	registered_at: string;
	last_refreshed: string;
}

export const load: PageServerLoad = async ({ fetch }) => {
	try {
		const res = await fetch(`${SYGIL_URL}/agents`, { signal: AbortSignal.timeout(3000) });
		if (!res.ok) throw new Error(`sygil ${res.status}`);
		const agents: SigilAgent[] = await res.json();
		return { sygilAgents: agents };
	} catch {
		return { sygilAgents: [] };
	}
};
