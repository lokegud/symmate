import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

const A2A_BACKEND = 'http://localhost:7843';
const MCP_BACKEND = 'http://localhost:7844';

async function proxyWellKnown(upstream: string): Promise<Response> {
	try {
		const res = await fetch(upstream);
		const body = await res.text();
		return new Response(body, {
			status: res.status,
			headers: {
				'content-type': 'application/json',
				'cache-control': 'max-age=300',
				'access-control-allow-origin': '*'
			}
		});
	} catch {
		return new Response(JSON.stringify({ error: 'backend unavailable' }), {
			status: 503,
			headers: { 'content-type': 'application/json' }
		});
	}
}

export const handle: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;

	// /.well-known/agent.json — legacy path; redirect to current spec path
	if (path === '/.well-known/agent.json') {
		throw redirect(301, '/.well-known/agent-card.json');
	}

	// /.well-known/agent-card.json — proxy to A2A server; 503 if A2A down (not vestibule 500)
	if (path === '/.well-known/agent-card.json') {
		return proxyWellKnown(A2A_BACKEND + path);
	}

	// /.well-known/mcp.json — proxy to MCP server; 503 if MCP down (not vestibule 500)
	if (path === '/.well-known/mcp.json') {
		return proxyWellKnown(MCP_BACKEND + path);
	}

	// /mcp/sse — proxy MCP SSE subscription endpoint (streaming, long-lived)
	if (path === '/mcp/sse') {
		try {
			const upstream = await fetch(MCP_BACKEND + '/sse');
			return new Response(upstream.body, {
				status: upstream.status,
				headers: {
					'content-type': 'text/event-stream',
					'cache-control': 'no-cache',
					'connection': 'keep-alive',
					'x-accel-buffering': 'no',
					'access-control-allow-origin': '*'
				}
			});
		} catch {
			return new Response('backend unavailable', { status: 503 });
		}
	}

	// /mcp/message — proxy MCP JSON-RPC message endpoint (preserves ?sessionId=...)
	if (path === '/mcp/message') {
		try {
			const upstreamURL = MCP_BACKEND + '/message' + event.url.search;
			const upstream = await fetch(upstreamURL, {
				method: event.request.method,
				headers: { 'content-type': 'application/json' },
				body: event.request.method !== 'GET' ? await event.request.text() : undefined
			});
			const body = await upstream.text();
			return new Response(body, {
				status: upstream.status,
				headers: { 'content-type': 'application/json' }
			});
		} catch {
			return new Response(JSON.stringify({ error: 'backend unavailable' }), {
				status: 503,
				headers: { 'content-type': 'application/json' }
			});
		}
	}

	// Owner routes are served by the operator app (port 7091), not the vestibule.
	// Return 404 so visitors cannot accidentally reach the operator surface.
	if (path.startsWith('/owner')) {
		return new Response('Not found', { status: 404 });
	}

	return resolve(event);
};
