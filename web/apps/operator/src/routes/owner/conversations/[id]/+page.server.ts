import { error } from '@sveltejs/kit';
import { ownerGetConversation } from '@symmate/api-client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, params }) => {
	const token = locals.ownerToken!;
	try {
		const conversation = await ownerGetConversation(token, params.id);
		return { conversation };
	} catch {
		throw error(404, 'Conversation not found');
	}
};
