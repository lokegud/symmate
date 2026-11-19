import { fail, redirect } from '@sveltejs/kit';
import { ownerLogin } from '@symmate/api-client';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	// Already logged in — go to dashboard.
	if (cookies.get('owner_token')) throw redirect(303, '/owner');
	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const phrase = data.get('phrase')?.toString().trim() ?? '';

		if (!phrase) return fail(400, { error: 'Phrase is required' });

		try {
			const { token } = await ownerLogin(phrase);
			cookies.set('owner_token', token, {
				httpOnly: true,
				secure: false, // set true when behind HTTPS
				sameSite: 'lax',
				path: '/',
				maxAge: 60 * 60 * 24 // 24h
			});
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : 'Login failed';
			return fail(401, { error: msg });
		}

		throw redirect(303, '/owner');
	}
};
