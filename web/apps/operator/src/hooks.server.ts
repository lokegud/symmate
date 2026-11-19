import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;

	// Redirect root to /owner dashboard.
	if (path === '/') {
		throw redirect(302, '/owner');
	}

	// Protect all /owner/* routes except login.
	if (path.startsWith('/owner') && !path.startsWith('/owner/login')) {
		const token = event.cookies.get('owner_token');
		if (!token) {
			throw redirect(303, '/owner/login');
		}
		event.locals.ownerToken = token;
	}

	return resolve(event);
};
