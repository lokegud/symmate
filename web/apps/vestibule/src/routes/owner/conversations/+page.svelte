<script lang="ts">
	import type { PageData } from './$types';
	import type { ConversationSummary } from '@symmate/api-client';

	let { data }: { data: PageData } = $props();

	const conversations: ConversationSummary[] = $derived(data.conversations ?? []);

	function fmtDate(ms: number) {
		return new Date(ms).toLocaleString('en-GB', {
			day: '2-digit', month: 'short', year: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	function duration(start: number, end?: number) {
		if (!end) return 'Active';
		const secs = Math.round((end - start) / 1000);
		if (secs < 60) return `${secs}s`;
		return `${Math.round(secs / 60)}m`;
	}
</script>

<svelte:head>
	<title>Conversations · Owner Dashboard</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1 class="font-display">Conversations</h1>
		<p class="subtitle">All visitor conversations, newest first.</p>
	</header>

	{#if conversations.length === 0}
		<div class="empty">
			<svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
			<p>No conversations yet.</p>
		</div>
	{:else}
		<div class="list">
			{#each conversations as conv}
				<a href="/owner/conversations/{conv.id}" class="conv-row">
					<div class="conv-meta">
						<span class="conv-handle">
							{conv.visitor_handle ?? 'Anonymous visitor'}
						</span>
						<span class="conv-date">{fmtDate(conv.started_at)}</span>
					</div>
					<div class="conv-details">
						<span class="conv-surface">{conv.surface}</span>
						<span class="conv-duration">{duration(conv.started_at, conv.ended_at)}</span>
						{#if conv.outcome}
							<span class="badge badge--{conv.outcome === 'completed' ? 'success' : 'neutral'}">{conv.outcome}</span>
						{/if}
					</div>
					<svg class="arrow" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polyline points="9 18 15 12 9 6"/></svg>
				</a>
			{/each}
		</div>
	{/if}
</div>

<style>
	.page {
		max-width: 800px;
		display: flex;
		flex-direction: column;
		gap: var(--space-8);
	}

	.page-header h1 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-3xl);
		font-weight: var(--weight-semibold);
		letter-spacing: -0.03em;
		color: var(--color-text);
	}

	.subtitle {
		margin-top: var(--space-2);
		font-size: var(--text-base);
		color: var(--color-text-muted);
	}

	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-4);
		padding: var(--space-16) 0;
		color: var(--color-text-subtle);
		text-align: center;
	}

	.list {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}

	.conv-row {
		display: flex;
		align-items: center;
		gap: var(--space-4);
		padding: var(--space-4) var(--space-5);
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-md);
		text-decoration: none;
		color: inherit;
		transition: border-color 100ms, box-shadow 100ms;
	}

	.conv-row:hover {
		border-color: var(--color-border);
		box-shadow: var(--shadow-sm);
	}

	.conv-meta {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: var(--space-0-5);
	}

	.conv-handle {
		font-size: var(--text-base);
		font-weight: var(--weight-medium);
		color: var(--color-text);
	}

	.conv-date {
		font-size: var(--text-xs);
		color: var(--color-text-muted);
		font-variant-numeric: tabular-nums;
	}

	.conv-details {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}

	.conv-surface {
		font-size: var(--text-xs);
		font-family: var(--font-mono);
		color: var(--color-text-subtle);
	}

	.conv-duration {
		font-size: var(--text-xs);
		color: var(--color-text-muted);
		font-variant-numeric: tabular-nums;
	}

	.badge {
		display: inline-flex;
		align-items: center;
		height: 20px;
		padding: 0 var(--space-1-5);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: var(--weight-semibold);
	}

	.badge--success { background: var(--color-success-subtle); color: var(--color-success); }
	.badge--neutral { background: var(--color-surface-raised); color: var(--color-text-muted); }

	.arrow {
		color: var(--color-text-subtle);
		flex-shrink: 0;
	}
</style>
