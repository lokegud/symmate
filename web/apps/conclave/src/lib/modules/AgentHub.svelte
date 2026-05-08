<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchCohortState, agentStore, HUB_URL, type CohortMember } from '../store.svelte.js';

	let members = $state<CohortMember[]>([]);
	let selectedId = $state<string | null>(null);
	let error = $state(false);

	function statusColor(status: string): string {
		if (status === 'working') return 'var(--state-working)';
		if (status === 'offline') return 'var(--state-revoked)';
		return 'var(--state-idle)';
	}

	function relativeTime(iso: string): string {
		const diff = Date.now() - new Date(iso).getTime();
		const mins = Math.floor(diff / 60000);
		if (mins < 1) return 'just now';
		if (mins < 60) return `${mins}m ago`;
		return `${Math.floor(mins / 60)}h ago`;
	}

	// Strip @user:server → user
	function shortSender(sender: string): string {
		return sender.replace(/^@/, '').split(':')[0];
	}

	onMount(() => {
		async function poll() {
			try {
				const data = await fetchCohortState();
				members = data;
				error = false;
			} catch {
				error = true;
			}
		}
		poll();
		const interval = setInterval(poll, 5000);
		return () => clearInterval(interval);
	});

	// Fall back to agentStore seed while first poll is in flight
	const display = $derived(
		members.length > 0 ? members : agentStore.agents.map(a => ({
			id: a.id,
			name: a.name,
			fingerprint: '',
			bindrune_url: '',
			status: (a.status === 'working' ? 'working' : 'idle') as CohortMember['status'],
			task: a.task,
			pane_cmd: '',
			last_seen: a.last_seen,
			source: 'sygil' as const,
		}))
	);
</script>

<div class="hub">
	{#if display.length === 0}
		<div class="empty">
			{error ? 'Hub unreachable — showing Sygil registry' : 'No agents registered'}
		</div>
	{:else}
		{#each display as m (m.id)}
			<div
				class="member"
				class:selected={selectedId === m.id}
				onclick={() => selectedId = selectedId === m.id ? null : m.id}
				role="button"
				tabindex="0"
				onkeydown={(e) => e.key === 'Enter' && (selectedId = m.id)}
			>
				<div class="member-header">
					{#if m.bindrune_url}
						<img class="bindrune" src={m.bindrune_url} alt={m.name} />
					{:else}
						<span class="bindrune-placeholder">{m.name[0]}</span>
					{/if}

					<div class="member-meta">
						<div class="name-row">
							<span class="member-name">{m.name}</span>
							<span class="source-badge" data-source={m.source}>{m.source}</span>
						</div>
						<div class="task-row">
							<span
								class="status-dot"
								style="background: {statusColor(m.status)}"
								title={m.status}
							></span>
							<span class="task-text">{m.task || m.status}</span>
						</div>
					</div>

					<span class="last-seen">{relativeTime(m.last_seen)}</span>
				</div>

				{#if selectedId === m.id && m.pane_cmd}
					<div class="pane-cmd">
						<span class="pane-label">pane</span>
						<code>{m.pane_cmd}</code>
					</div>
				{/if}
			</div>
		{/each}
	{/if}
</div>

<style>
	.hub {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: var(--hub-space-2);
		height: 100%;
		overflow-y: auto;
	}

	.empty {
		color: var(--hub-text-subtle);
		font-size: var(--hub-text-xs);
		text-align: center;
		padding: 32px 0;
		font-family: var(--hub-font-mono);
	}

	.member {
		padding: var(--hub-space-2) var(--hub-space-3);
		border-radius: var(--hub-radius-md);
		cursor: pointer;
		background: rgba(255,255,255,0.03);
		border: 1px solid var(--hub-border-subtle);
		transition: background var(--hub-duration-fast) var(--hub-ease);
	}
	.member:hover { background: rgba(255,255,255,0.055); }
	.member.selected { background: rgba(255,255,255,0.065); border-color: var(--hub-border); }

	.member-header {
		display: flex;
		align-items: center;
		gap: var(--hub-space-2);
	}

	.bindrune {
		width: 28px;
		height: 28px;
		border-radius: var(--hub-radius-sm);
		flex-shrink: 0;
		opacity: 0.85;
	}

	.bindrune-placeholder {
		width: 28px;
		height: 28px;
		border-radius: var(--hub-radius-sm);
		background: rgba(255,255,255,0.06);
		border: 1px solid var(--hub-border);
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: var(--hub-font-mono);
		font-size: var(--hub-text-xs);
		color: var(--hub-text-muted);
		flex-shrink: 0;
		font-weight: 600;
		text-transform: uppercase;
	}

	.member-meta {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.name-row {
		display: flex;
		align-items: center;
		gap: var(--hub-space-2);
	}

	.member-name {
		font-family: var(--hub-font-mono);
		font-size: var(--hub-text-sm);
		font-weight: 600;
		color: var(--hub-text);
	}

	.source-badge {
		font-size: var(--hub-text-2xs);
		font-family: var(--hub-font-mono);
		color: var(--hub-text-subtle);
		background: rgba(255,255,255,0.04);
		border: 1px solid var(--hub-border-subtle);
		border-radius: var(--hub-radius-sm);
		padding: 1px 4px;
		letter-spacing: 0.04em;
	}
	.source-badge[data-source="both"] {
		color: var(--hub-accent);
		border-color: var(--hub-accent-border);
		background: var(--hub-accent-dim);
	}

	.task-row {
		display: flex;
		align-items: center;
		gap: var(--hub-space-1);
		min-width: 0;
	}

	.status-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.task-text {
		font-size: var(--hub-text-xs);
		color: var(--hub-text-muted);
		font-family: var(--hub-font-mono);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.last-seen {
		font-size: var(--hub-text-2xs);
		font-family: var(--hub-font-mono);
		color: var(--hub-text-subtle);
		flex-shrink: 0;
	}

	.pane-cmd {
		display: flex;
		align-items: baseline;
		gap: var(--hub-space-2);
		margin-top: var(--hub-space-2);
		padding: var(--hub-space-1) var(--hub-space-2);
		background: rgba(0,0,0,0.3);
		border-radius: var(--hub-radius-sm);
		overflow: hidden;
	}

	.pane-label {
		font-size: var(--hub-text-2xs);
		font-family: var(--hub-font-mono);
		color: var(--hub-text-subtle);
		flex-shrink: 0;
	}

	code {
		font-family: var(--hub-font-mono);
		font-size: var(--hub-text-2xs);
		color: var(--hub-text-muted);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
	}
</style>
