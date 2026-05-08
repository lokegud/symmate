import { ownerListConversations } from '@symmate/api-client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	const token = locals.ownerToken!;
	const page = parseInt(url.searchParams.get('page') ?? '1');
	const result = await ownerListConversations(token, page).catch(() => ({
		conversations: [],
		page: 1
	}));
	return { ...result, page };
};
