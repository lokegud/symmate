import { ownerGetAudit } from '@symmate/api-client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	const token = locals.ownerToken!;
	const page = parseInt(url.searchParams.get('page') ?? '1');
	const audit = await ownerGetAudit(token, page).catch(() => ({
		entries: [],
		page: 1,
		total: 0,
		pages: 0
	}));
	return { audit, page };
};
