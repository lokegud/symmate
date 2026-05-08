import { fail } from '@sveltejs/kit';
import { ownerGetPolicy, ownerPutPolicy } from '@symmate/api-client';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const token = locals.ownerToken!;
	const policy = await ownerGetPolicy(token);
	return { policy };
};

export const actions: Actions = {
	default: async ({ request, locals }) => {
		const token = locals.ownerToken!;
		const data = await request.formData();
		const raw = data.get('tool_tags')?.toString() ?? '';
		const tool_tags = raw
			.split('\n')
			.map((t) => t.trim())
			.filter(Boolean);

		try {
			await ownerPutPolicy(token, { tool_tags });
			return { success: true };
		} catch (err: unknown) {
			return fail(500, { error: err instanceof Error ? err.message : 'Save failed' });
		}
	}
};
