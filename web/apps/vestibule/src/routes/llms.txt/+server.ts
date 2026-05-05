import { visitorBase } from '@symmate/api-client';
import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () => {
	const res = await fetch(`${visitorBase()}/llms.txt`);
	if (!res.ok) throw error(502, 'upstream unavailable');

	const text = await res.text();
	return new Response(text, {
		headers: {
			'Content-Type': 'text/plain; charset=utf-8',
			'Cache-Control': 'public, max-age=300'
		}
	});
};
