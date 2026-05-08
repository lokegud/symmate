<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const providerLabel: Record<string, string> = {
		anthropic: 'Anthropic',
		hermes: 'Hermes (local)',
		'openai-compat': 'OpenAI-compat'
	};
</script>

<svelte:head>
	<title>Dashboard · {data.summary.name}</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1 class="font-display">{data.summary.name}</h1>
		<p class="subtitle">Your Symmate is {data.summary.enrolled ? 'enrolled and active' : 'not yet enrolled'}.</p>
	</header>

	<div class="stat-grid">
		<div class="stat-card">
			<span class="stat-label">Conversations</span>
			<span class="stat-value">{data.summary.conversation_count}</span>
		</div>
		<div class="stat-card">
			<span class="stat-label">Enrolled devices</span>
			<span class="stat-value">{data.summary.devices}</span>
		</div>
		<div class="stat-card">
			<span class="stat-label">Model</span>
			<span class="stat-value stat-value--sm">
				{data.llmConfig ? `${providerLabel[data.llmConfig.provider] ?? data.llmConfig.provider} · ${data.llmConfig.model}` : 'Default'}
			</span>
			<a href="/owner/personality" class="stat-action">Change →</a>
		</div>
		<div class="stat-card">
			<span class="stat-label">Status</span>
			<span class="stat-value stat-value--badge {data.summary.enrolled ? 'badge--success' : 'badge--warning'}">
				{data.summary.enrolled ? 'Active' : 'Unenrolled'}
			</span>
		</div>
	</div>

	<div class="quick-links">
		<h2>Quick actions</h2>
		<div class="link-grid">
			<a href="/owner/personality" class="quick-card">
				<span class="qc-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M20 21a8 8 0 1 0-16 0"/></svg>
				</span>
				<span class="qc-title">Edit Personality</span>
				<span class="qc-desc">Update context and vertical behaviour.</span>
			</a>
			<a href="/owner/policy" class="quick-card">
				<span class="qc-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
				</span>
				<span class="qc-title">Set Policy</span>
				<span class="qc-desc">Configure tool grants and access rules.</span>
			</a>
			<a href="/owner/audit" class="quick-card">
				<span class="qc-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
				</span>
				<span class="qc-title">View Audit Log</span>
				<span class="qc-desc">Review every disclosure and tool call.</span>
			</a>
			<a href="/owner/conversations" class="quick-card">
				<span class="qc-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
				</span>
				<span class="qc-title">Conversations</span>
				<span class="qc-desc">Review visitor conversations.</span>
			</a>
		</div>
	</div>
</div>

<style>
	.page {
		max-width: 900px;
		display: flex;
		flex-direction: column;
		gap: var(--space-10);
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

	/* Stats */
	.stat-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: var(--space-4);
	}

	.stat-card {
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-lg);
		padding: var(--space-6);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		box-shadow: var(--shadow-sm);
	}

	.stat-label {
		font-size: var(--text-xs);
		font-weight: var(--weight-semibold);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--color-text-muted);
	}

	.stat-value {
		font-size: var(--text-3xl);
		font-weight: var(--weight-bold);
		color: var(--color-text);
		font-variant-numeric: tabular-nums;
	}

	.stat-value--sm {
		font-size: var(--text-base);
		font-weight: var(--weight-medium);
	}

	.stat-value--badge {
		display: inline-flex;
		align-items: center;
		height: 24px;
		padding: 0 var(--space-2);
		border-radius: var(--radius-sm);
		font-size: var(--text-sm);
		font-weight: var(--weight-semibold);
		letter-spacing: 0.04em;
		width: fit-content;
	}

	.badge--success { background: var(--color-success-subtle); color: var(--color-success); }
	.badge--warning { background: var(--color-warning-subtle); color: var(--color-warning); }

	.stat-action {
		font-size: var(--text-xs);
		color: var(--color-accent);
		text-decoration: none;
		margin-top: auto;
	}

	.stat-action:hover { text-decoration: underline; }

	/* Quick links */
	.quick-links h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		margin-bottom: var(--space-4);
	}

	.link-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-4);
	}

	.quick-card {
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-lg);
		padding: var(--space-6);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		text-decoration: none;
		color: inherit;
		transition: border-color 150ms, box-shadow 150ms, transform 150ms;
	}

	.quick-card:hover {
		border-color: var(--color-border);
		box-shadow: var(--shadow-md);
		transform: translateY(-1px);
	}

	.qc-icon {
		color: var(--color-accent);
		display: flex;
	}

	.qc-title {
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}

	.qc-desc {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		line-height: 1.5;
	}
</style>
