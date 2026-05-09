import { redirect } from '@sveltejs/kit';
import { ownerLogout } from '@symmate/api-client';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async ({ cookies }) => {
		const token = cookies.get('owner_token');
		if (token) {
			// Best-effort logout on the API side.
			await ownerLogout(token).catch(() => {});
		}
		cookies.delete('owner_token', { path: '/' });
		throw redirect(303, '/owner/login');
	}
};
