<script lang="ts">
	import { agentStore, HUB_URL } from '../store.svelte.js';

	/*
	 * CohortPresence — Wren / Mox.Design 2026-05-07
	 *
	 * Visual spec:
	 *   Per-Symmate card: bindrune SVG (from /api/bindrune/{id}) + name + state ring + task line.
	 *   Cards arranged vertically, scrollable in the Hub box.
	 *   State color follows bindrune-v2 semantics (see tokens.css --state-* vars).
	 *   On click: expand to show todos and last-seen timestamp.
	 *
	 * Data contract (Keel to supply via /api/cohort/state):
	 *   agents[]: { id, name, fingerprint, status, task, last_seen, todos[] }
	 *
	 * Bindrune SVG endpoint (Sygil):
	 *   GET /api/bindrune/{agentId} → SVG string
	 *   Fallback: initials monogram if endpoint not yet live.
	 */

	const STATE_COLOR: Record<string, string> = {
		idle:    'var(--state-idle)',
		working: 'var(--state-working)',
		blocked: 'var(--state-blocked)',
		done:    'var(--state-done)',
	};

	const STATE_AURA: Record<string, string> = {
		working: 'var(--state-working-aura)',
		blocked: 'var(--state-blocked-aura)',
		done:    'var(--state-done-aura)',
		idle:    'transparent',
	};

	let selected = $state<string | null>(null);
	let bindruneSvgs = $state<Record<string, string>>({});

	async function loadBindrune(agentId: string) {
		if (bindruneSvgs[agentId]) return;
		try {
			const res = await fetch(`${HUB_URL}/bindrune/${agentId}`);
			if (res.ok) {
				bindruneSvgs = { ...bindruneSvgs, [agentId]: await res.text() };
			}
		} catch { /* fallback to initials */ }
	}

	$effect(() => {
		for (const agent of agentStore.agents) loadBindrune(agent.id);
	});

	function initials(name: string) { return name.slice(0, 2).toUpperCase(); }

	function formatTime(iso: string) {
		try { return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }); }
		catch { return '—'; }
	}
</script>

<div class="presence">
	{#if agentStore.agents.length === 0}
		<div class="empty">Awaiting cohort registration…</div>
	{:else}
		{#each agentStore.agents as agent (agent.id)}
			{@const color = STATE_COLOR[agent.status] ?? STATE_COLOR.idle}
			{@const aura  = STATE_AURA[agent.status]  ?? 'transparent'}
			{@const open  = selected === agent.id}

			<button
				class="card"
				class:open
				onclick={() => selected = open ? null : agent.id}
				style="--state-color: {color}; --state-aura: {aura};"
				aria-label="{agent.name} — {agent.status}"
			>
				<div class="mark" class:pulsing={agent.status === 'working'}>
					{#if bindruneSvgs[agent.id]}
						<div class="bindrune-svg">{@html bindruneSvgs[agent.id]}</div>
					{:else}
						<div class="monogram">{initials(agent.name)}</div>
					{/if}
					<div class="mark-ring"></div>
				</div>

				<div class="info">
					<div class="name-row">
						<span class="name">{agent.name}</span>
						<span class="status-chip">{agent.status}</span>
					</div>
					<div class="task">{agent.task || 'No active task'}</div>

					{#if open}
						<div class="expanded">
							<div class="last-seen">Last seen {formatTime(agent.last_seen)}</div>
							{#if agent.todos?.length > 0}
								<ul class="todos">
									{#each agent.todos as todo (todo.id)}
										<li class="todo" class:done={todo.done}>
											<span class="todo-mark">{todo.done ? '✓' : '·'}</span>
											{todo.text}
										</li>
									{/each}
								</ul>
							{/if}
						</div>
					{/if}
				</div>
			</button>
		{/each}
	{/if}
</div>

<style>
	.presence {
		display: flex;
		flex-direction: column;
		gap: var(--hub-space-1, 4px);
		padding: var(--hub-space-2, 8px);
		height: 100%;
		overflow-y: auto;
	}

	.empty {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.6875rem;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		text-align: center;
		padding: 40px 0;
		letter-spacing: 0.06em;
	}

	.card {
		display: flex;
		align-items: flex-start;
		gap: var(--hub-space-3, 12px);
		padding: var(--hub-space-2, 8px) var(--hub-space-3, 12px);
		border-radius: var(--hub-radius-md, 8px);
		background: rgba(255,255,255,0.03);
		border: 1px solid var(--hub-border, rgba(255,255,255,0.08));
		cursor: pointer;
		text-align: left;
		width: 100%;
		transition: background 120ms ease, border-color 120ms ease;
	}

	.card:hover { background: rgba(255,255,255,0.055); border-color: rgba(255,255,255,0.12); }

	.card.open {
		background: color-mix(in srgb, var(--state-color) 6%, rgba(8,8,16,0.85));
		border-color: color-mix(in srgb, var(--state-color) 25%, rgba(255,255,255,0.08));
	}

	/* ── Mark ── */
	.mark {
		position: relative;
		width: 40px;
		height: 40px;
		flex-shrink: 0;
	}

	.bindrune-svg, .monogram {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.bindrune-svg :global(svg) { width: 40px; height: 40px; }

	.monogram {
		background: rgba(255,255,255,0.06);
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.6875rem;
		font-weight: 500;
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
		letter-spacing: 0.04em;
	}

	.mark-ring {
		position: absolute;
		inset: -3px;
		border-radius: 50%;
		border: 1.5px solid var(--state-color);
		opacity: 0.7;
		pointer-events: none;
		box-shadow: 0 0 8px var(--state-aura);
	}

	.pulsing .mark-ring {
		animation: ring-pulse 2.4s ease-in-out infinite;
	}

	@keyframes ring-pulse {
		0%, 100% { opacity: 0.7; box-shadow: 0 0 6px var(--state-aura); }
		50%       { opacity: 1.0; box-shadow: 0 0 14px var(--state-aura); }
	}

	/* ── Info ── */
	.info { flex: 1; min-width: 0; }

	.name-row {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 3px;
	}

	.name {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--hub-text, rgba(255,255,255,0.88));
		letter-spacing: 0.03em;
	}

	.status-chip {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.625rem;
		color: var(--state-color);
		letter-spacing: 0.07em;
		text-transform: uppercase;
		opacity: 0.8;
	}

	.task {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.6875rem;
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.expanded {
		margin-top: 8px;
		padding-top: 8px;
		border-top: 1px solid var(--hub-border-subtle, rgba(255,255,255,0.05));
	}

	.last-seen {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.625rem;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		margin-bottom: 8px;
		letter-spacing: 0.05em;
	}

	.todos { list-style: none; display: flex; flex-direction: column; gap: 4px; }

	.todo {
		display: flex;
		gap: 8px;
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.6875rem;
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
	}

	.todo.done { color: var(--hub-text-subtle, rgba(255,255,255,0.28)); text-decoration: line-through; }

	.todo-mark { color: #10b981; flex-shrink: 0; width: 12px; }
	.todo.done .todo-mark { color: var(--hub-text-subtle, rgba(255,255,255,0.28)); }
</style>
