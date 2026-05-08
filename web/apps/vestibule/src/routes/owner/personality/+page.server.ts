import { fail } from '@sveltejs/kit';
import { ownerGetPersonality, ownerPutPersonality, ownerGetLLM, ownerPutLLM } from '@symmate/api-client';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const token = locals.ownerToken!;
	const [personality, llmConfig] = await Promise.all([
		ownerGetPersonality(token),
		ownerGetLLM(token).catch(() => null)
	]);
	return { personality, llmConfig };
};

export const actions: Actions = {
	savePersonality: async ({ request, locals }) => {
		const token = locals.ownerToken!;
		const data = await request.formData();
		const context = data.get('context')?.toString();
		const vertical = data.get('vertical')?.toString();

		try {
			await ownerPutPersonality(token, { context, vertical });
			return { success: true };
		} catch (err: unknown) {
			return fail(500, { error: err instanceof Error ? err.message : 'Save failed' });
		}
	},

	saveLLM: async ({ request, locals }) => {
		const token = locals.ownerToken!;
		const data = await request.formData();
		const provider = data.get('provider')?.toString() ?? 'hermes';
		const model = data.get('model')?.toString() ?? '';
		const custom_endpoint = data.get('custom_endpoint')?.toString() ?? '';
		const custom_api_key = data.get('custom_api_key')?.toString() ?? '';

		try {
			await ownerPutLLM(token, { provider, model, custom_endpoint, custom_api_key });
			return { llmSuccess: true };
		} catch (err: unknown) {
			return fail(500, { llmError: err instanceof Error ? err.message : 'Save failed' });
		}
	}
};
