import { ownerSummary, ownerGetLLM } from '@symmate/api-client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const token = locals.ownerToken!;
	const [summary, llmConfig] = await Promise.all([
		ownerSummary(token),
		ownerGetLLM(token).catch(() => null)
	]);
	return { summary, llmConfig };
};
