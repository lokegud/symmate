import { fail } from '@sveltejs/kit';
import {
	ownerCreateUser,
	ownerDeleteUser,
	ownerListUsers,
	ownerRotateUserToken,
	ownerSetUsersEnabled,
	ownerUpdateUser
} from '@symmate/api-client';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const users = await ownerListUsers(locals.ownerToken!);
	return { users };
};

function message(err: unknown): string {
	return err instanceof Error ? err.message : 'Request failed';
}

export const actions: Actions = {
	settings: async ({ request, locals }) => {
		const data = await request.formData();
		try {
			await ownerSetUsersEnabled(locals.ownerToken!, data.get('enabled') === 'true');
			return { success: 'Saved.' };
		} catch (err) {
			return fail(500, { error: message(err) });
		}
	},

	create: async ({ request, locals }) => {
		const data = await request.formData();
		const name = data.get('name')?.toString().trim() ?? '';
		const groups = data.getAll('groups').map(String);
		if (!name) return fail(400, { error: 'Name is required.' });
		try {
			const res = await ownerCreateUser(locals.ownerToken!, name, groups);
			return { success: `Added ${res.user.name}.`, newToken: res.token, tokenFor: res.user.name };
		} catch (err) {
			return fail(400, { error: message(err) });
		}
	},

	update: async ({ request, locals }) => {
		const data = await request.formData();
		const id = data.get('id')?.toString() ?? '';
		try {
			await ownerUpdateUser(locals.ownerToken!, id, {
				enabled: data.get('enabled') === 'true',
				groups: data.getAll('groups').map(String)
			});
			return { success: 'Saved.' };
		} catch (err) {
			return fail(400, { error: message(err) });
		}
	},

	rotate: async ({ request, locals }) => {
		const data = await request.formData();
		const id = data.get('id')?.toString() ?? '';
		const name = data.get('name')?.toString() ?? '';
		try {
			const res = await ownerRotateUserToken(locals.ownerToken!, id);
			return { success: `New token issued for ${name}. The old one no longer works.`, newToken: res.token, tokenFor: name };
		} catch (err) {
			return fail(400, { error: message(err) });
		}
	},

	remove: async ({ request, locals }) => {
		const data = await request.formData();
		const id = data.get('id')?.toString() ?? '';
		try {
			await ownerDeleteUser(locals.ownerToken!, id);
			return { success: 'User removed.' };
		} catch (err) {
			return fail(400, { error: message(err) });
		}
	}
};
