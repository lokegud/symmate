import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

// Owner routes require a valid owner_token cookie.
// Token is minted by the Go owner API on POST /auth/login and stored in the cookie.
export const handle: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;

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
