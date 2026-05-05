import { getAgentCard } from '@symmate/api-client';
import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () => {
	try {
		const card = await getAgentCard();
		return json(card, {
			headers: { 'Cache-Control': 'public, max-age=300' }
		});
	} catch (e) {
		throw error(502, 'upstream unavailable');
	}
};
