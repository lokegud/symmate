import { getPersonaIdentity } from '@symmate/api-client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const persona = await getPersonaIdentity();
	return { persona };
};
