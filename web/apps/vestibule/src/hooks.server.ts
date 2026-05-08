import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

const A2A_BACKEND = 'http://localhost:7843';
const MCP_BACKEND = 'http://localhost:7844';

export const handle: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;

	// /.well-known/agent.json — legacy path; redirect to current spec path
	if (path === '/.well-known/agent.json') {
		throw redirect(301, '/.well-known/agent-card.json');
	}

	// /.well-known/agent-card.json — proxy to A2A server
	if (path === '/.well-known/agent-card.json') {
		const upstream = await fetch(A2A_BACKEND + path);
		const body = await upstream.text();
		return new Response(body, {
			status: upstream.status,
			headers: {
				'content-type': 'application/json',
				'cache-control': 'max-age=300',
				'access-control-allow-origin': '*'
			}
		});
	}

	// /.well-known/mcp.json — proxy to MCP server
	if (path === '/.well-known/mcp.json') {
		const upstream = await fetch(MCP_BACKEND + path);
		const body = await upstream.text();
		return new Response(body, {
			status: upstream.status,
			headers: {
				'content-type': 'application/json',
				'cache-control': 'max-age=300',
				'access-control-allow-origin': '*'
			}
		});
	}

	// Owner routes are served by the operator app (port 7091), not the vestibule.
	// Return 404 so visitors cannot accidentally reach the operator surface.
	if (path.startsWith('/owner')) {
		return new Response('Not found', { status: 404 });
	}

	return resolve(event);
};
