import { startChatStream } from '@symmate/api-client';
import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';

// Proxy POST /chat to runtime :7842/chat and pass SSE stream through.
export const POST: RequestHandler = async ({ request }) => {
	let body: { message: string };
	try {
		body = await request.json();
	} catch {
		throw error(400, 'invalid json');
	}

	if (!body.message?.trim()) {
		throw error(400, 'message required');
	}

	const upstream = await startChatStream(body.message);

	if (!upstream.ok) {
		throw error(upstream.status, 'upstream error');
	}

	return new Response(upstream.body, {
		headers: {
			'Content-Type': 'text/event-stream',
			'Cache-Control': 'no-cache',
			Connection: 'keep-alive',
			'X-Accel-Buffering': 'no'
		}
	});
};
