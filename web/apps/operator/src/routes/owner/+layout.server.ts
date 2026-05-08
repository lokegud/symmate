import { redirect } from '@sveltejs/kit';
import { ownerSummary } from '@symmate/api-client';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals, url }) => {
	// Login page doesn't need the token check.
	if (url.pathname === '/owner/login') return {};

	const token = locals.ownerToken;
	if (!token) throw redirect(303, '/owner/login');

	// Validate token by hitting /summary.
	let summary;
	try {
		summary = await ownerSummary(token);
	} catch {
		throw redirect(303, '/owner/login');
	}

	return { summary, token };
};
