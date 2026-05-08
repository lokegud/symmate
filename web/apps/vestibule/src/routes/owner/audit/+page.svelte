<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	type AuditEntry = {
		id: string;
		timestamp: string;
		category: string;
		severity: string;
		conversation_id?: string;
		payload: Record<string, unknown>;
	};

	const entries: AuditEntry[] = $derived(data.audit.entries as AuditEntry[]);

	const severityClass: Record<string, string> = {
		info: 'badge--info',
		notice: 'badge--notice',
		warning: 'badge--warning',
		alert: 'badge--alert'
	};

	function fmt(ts: string) {
		return new Date(ts).toLocaleString('en-GB', {
			day: '2-digit', month: 'short', year: 'numeric',
			hour: '2-digit', minute: '2-digit', second: '2-digit'
		});
	}

	let expandedId = $state<string | null>(null);
	function toggle(id: string) {
		expandedId = expandedId === id ? null : id;
	}
</script>

<svelte:head>
	<title>Audit Log · Owner Dashboard</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1 class="font-display">Audit Log</h1>
		<p class="subtitle">Every disclosure, tool call, and firewall decision — most recent first.</p>
		{#if data.audit.total > 0}
			<span class="total-badge">{data.audit.total} entries</span>
		{/if}
	</header>

	{#if entries.length === 0}
		<div class="empty">
			<svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
			<p>No audit entries yet.</p>
		</div>
	{:else}
		<div class="log">
			{#each entries as entry}
				<div class="log-entry">
					<button
						class="log-header"
						onclick={() => toggle(entry.id)}
						aria-expanded={expandedId === entry.id}
					>
						<span class="log-time font-mono">{fmt(entry.timestamp)}</span>
						<span class="log-category">{entry.category}</span>
						<span class="badge {severityClass[entry.severity] ?? 'badge--neutral'}">{entry.severity}</span>
						{#if entry.conversation_id}
							<span class="conv-id font-mono">{entry.conversation_id.slice(0, 8)}</span>
						{/if}
						<svg
							class="chevron"
							class:chevron--open={expandedId === entry.id}
							width="14"
							height="14"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							aria-hidden="true"
						><polyline points="6 9 12 15 18 9"/></svg>
					</button>

					{#if expandedId === entry.id}
						<pre class="log-payload font-mono">{JSON.stringify(entry.payload, null, 2)}</pre>
					{/if}
				</div>
			{/each}
		</div>

		{#if data.audit.pages > 1}
			<nav class="pagination" aria-label="Audit log pages">
				{#each Array.from({ length: data.audit.pages }, (_, i) => i + 1) as p}
					<a
						href="?page={p}"
						class="page-btn"
						class:page-btn--active={p === data.page}
						aria-current={p === data.page ? 'page' : undefined}
					>{p}</a>
				{/each}
			</nav>
		{/if}
	{/if}
</div>

<style>
	.page {
		max-width: 900px;
		display: flex;
		flex-direction: column;
		gap: var(--space-8);
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
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
		font-size: var(--text-base);
		color: var(--color-text-muted);
	}

	.total-badge {
		display: inline-flex;
		align-items: center;
		height: 22px;
		padding: 0 var(--space-2);
		background: var(--color-surface-raised);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: var(--weight-semibold);
		color: var(--color-text-muted);
		letter-spacing: 0.04em;
		width: fit-content;
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

	.log {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}

	.log-entry {
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	.log-header {
		width: 100%;
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-3) var(--space-4);
		background: none;
		border: none;
		cursor: pointer;
		font-family: var(--font-body);
		text-align: left;
		transition: background-color 100ms;
		flex-wrap: wrap;
	}

	.log-header:hover { background: var(--color-surface-raised); }

	.log-time {
		font-size: var(--text-xs);
		color: var(--color-text-muted);
		white-space: nowrap;
		font-family: var(--font-mono);
	}

	.log-category {
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		flex: 1;
	}

	.conv-id {
		font-size: var(--text-xs);
		font-family: var(--font-mono);
		color: var(--color-text-subtle);
	}

	.chevron {
		color: var(--color-text-subtle);
		transition: transform 150ms;
		flex-shrink: 0;
		margin-left: auto;
	}

	.chevron--open { transform: rotate(180deg); }

	.badge {
		display: inline-flex;
		align-items: center;
		height: 20px;
		padding: 0 var(--space-1-5);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: var(--weight-semibold);
		letter-spacing: 0.04em;
		text-transform: uppercase;
		white-space: nowrap;
	}

	.badge--info    { background: var(--color-accent-subtle);  color: var(--color-accent); }
	.badge--notice  { background: var(--color-success-subtle); color: var(--color-success); }
	.badge--warning { background: var(--color-warning-subtle); color: var(--color-warning); }
	.badge--alert   { background: var(--color-error-subtle);   color: var(--color-error); }
	.badge--neutral { background: var(--color-surface-raised); color: var(--color-text-muted); }

	.log-payload {
		background: var(--color-surface-raised);
		border-top: 1px solid var(--color-border-subtle);
		padding: var(--space-4);
		margin: 0;
		font-size: var(--text-xs);
		font-family: var(--font-mono);
		color: var(--color-text-muted);
		overflow-x: auto;
		line-height: 1.6;
		white-space: pre-wrap;
		word-break: break-all;
	}

	.pagination {
		display: flex;
		gap: var(--space-1);
		flex-wrap: wrap;
	}

	.page-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: var(--radius-md);
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		color: var(--color-text-muted);
		text-decoration: none;
		transition: background-color 100ms, color 100ms;
	}

	.page-btn:hover { background: var(--color-surface-raised); color: var(--color-text); }

	.page-btn--active {
		background: var(--color-accent);
		color: var(--color-accent-text);
		border-color: var(--color-accent);
	}
</style>
