<script lang="ts">
	import { onMount } from 'svelte';
	import { HUB_URL } from '../store.svelte.js';

	/*
	 * TaskBoard — Wren / Mox.Design 2026-05-07
	 *
	 * Round-robin state visualizer.
	 * Pattern: ATC sector ownership map + NASA MOCR role assignment board.
	 *
	 * Each sibling is a hexagonal node in the chain.
	 * Current slot: bright, active. Past: faded. Future: outline only.
	 * PAUSE: amber stripe replaces connector lines.
	 *
	 * Data: GET /api/conclave/round (Keel)
	 * Fallback: derives state from /api/conclave/messages if round endpoint
	 * not yet live.
	 */

	const ROUND_ORDER = ['elder', 'keel', 'quill', 'trace', 'wren', 'wright'] as const;
	type SlotName = typeof ROUND_ORDER[number];

	interface RoundSlot {
		name: SlotName;
		agent_id: string;
		status: 'pending' | 'active' | 'complete';
		task: string;
		started_at: string | null;
	}

	interface RoundState {
		round: number;
		current_slot: SlotName;
		pause_active: boolean;
		pause_posted_by: string | null;
		slots: RoundSlot[];
	}

	let roundState: RoundState | null = $state(null);
	let error: boolean = $state(false);

	// Slot color accents
	const SLOT_ACCENT: Record<SlotName, string> = {
		elder:  '#e8a878',
		keel:   '#4a9eff',
		quill:  '#8b5cf6',
		trace:  '#f59e0b',
		wren:   '#10b981',
		wright: '#3d8ef0',
	};

	function elapsed(iso: string | null): string {
		if (!iso) return '';
		const mins = Math.floor((Date.now() - new Date(iso).getTime()) / 60000);
		if (mins < 1) return 'just now';
		if (mins < 60) return `${mins}m`;
		return `${Math.floor(mins / 60)}h ${mins % 60}m`;
	}

	function slotStatus(slot: RoundSlot, currentSlot: SlotName): 'past' | 'active' | 'future' {
		const currentIdx = ROUND_ORDER.indexOf(currentSlot);
		const slotIdx = ROUND_ORDER.indexOf(slot.name);
		if (slotIdx < currentIdx) return 'past';
		if (slotIdx === currentIdx) return 'active';
		return 'future';
	}

	// Fallback: derive rough state from recent messages if /api/conclave/round not live
	async function fetchRoundState(): Promise<RoundState> {
		try {
			const r = await fetch(`${HUB_URL}/conclave/round`);
			if (r.ok) return r.json();
		} catch { /* fall through to fallback */ }

		// Fallback: assume current slot from most recent round-assignment-language in messages
		return {
			round: 0,
			current_slot: 'elder',
			pause_active: false,
			pause_posted_by: null,
			slots: ROUND_ORDER.map(name => ({
				name,
				agent_id: name,
				status: 'pending' as const,
				task: '',
				started_at: null,
			})),
		};
	}

	onMount(() => {
		async function poll() {
			try {
				roundState = await fetchRoundState();
				error = false;
			} catch { error = true; }
		}
		poll();
		const iv = setInterval(poll, 5000);
		return () => clearInterval(iv);
	});

	function getActiveSlot(rs: RoundState | null): RoundSlot | null {
		if (!rs) return null;
		return rs.slots.find((s: RoundSlot) => s.name === rs.current_slot) ?? null;
	}

	const activeSlot: RoundSlot | null = $derived(getActiveSlot(roundState));
</script>

<div class="board">
	<!-- Round number -->
	{#if roundState}
		<div class="round-label">
			{roundState.pause_active
				? `⚠ PAUSE — ${roundState.pause_posted_by ?? 'unknown'}`
				: `Round ${roundState.round || '—'}`}
		</div>
	{/if}

	<!-- Hex chain -->
	<div class="chain" class:paused={roundState?.pause_active}>
		{#if roundState}
			{#each ROUND_ORDER as name, i}
				{@const slot = roundState.slots.find((s: RoundSlot) => s.name === name)}
				{@const rel  = slot ? slotStatus(slot, roundState.current_slot) : 'future'}
				{@const accent = SLOT_ACCENT[name]}

				{#if i > 0}
					<div class="connector" class:paused={roundState.pause_active}></div>
				{/if}

				<div
					class="hex-node"
					data-rel={rel}
					style="--accent: {accent}"
					title="{name}{slot?.task ? ' — ' + slot.task : ''}"
				>
					<div class="hex-inner">
						<span class="hex-name">{name}</span>
						{#if rel === 'active'}
							<span class="active-dot"></span>
						{/if}
					</div>
				</div>
			{/each}

			<!-- Loop indicator -->
			<div class="connector loop"></div>
			<div class="loop-mark">↺</div>
		{:else}
			<div class="loading">Syncing round state…</div>
		{/if}
	</div>

	<!-- Active slot detail -->
	{#if activeSlot && roundState}
		<div class="active-detail" style="--accent: {SLOT_ACCENT[activeSlot.name]}">
			<span class="active-who">{activeSlot.name}</span>
			<span class="active-sep">—</span>
			<span class="active-task">{activeSlot.task || roundState.pause_active ? 'holding' : 'no task set'}</span>
			{#if activeSlot.started_at}
				<span class="active-elapsed">{elapsed(activeSlot.started_at)}</span>
			{/if}
		</div>
	{/if}

	{#if error}
		<div class="err">Round endpoint unreachable — waiting</div>
	{/if}
</div>

<style>
	.board {
		display: flex;
		flex-direction: column;
		padding: var(--hub-space-3, 12px);
		gap: var(--hub-space-3, 12px);
		height: 100%;
		font-family: var(--hub-font-mono, ui-monospace, monospace);
	}

	.round-label {
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		letter-spacing: 0.08em;
		text-transform: uppercase;
		align-self: flex-end;
	}

	/* ── Hex chain ── */
	.chain {
		display: flex;
		align-items: center;
		gap: 0;
		overflow-x: auto;
		padding: 4px 0;
	}

	.chain.paused .connector { border-color: var(--state-blocked, #f59e0b); }

	.connector {
		flex: 1;
		height: 1px;
		border-top: 1px dashed var(--hub-border, rgba(255,255,255,0.08));
		min-width: 8px;
		max-width: 24px;
	}

	.connector.paused { border-color: var(--state-blocked, #f59e0b); opacity: 0.5; }
	.connector.loop { border-style: dotted; opacity: 0.4; }

	.loop-mark {
		font-size: 14px;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		flex-shrink: 0;
	}

	/* ── Hex node ── */
	.hex-node {
		flex-shrink: 0;
		width: 52px;
		height: 52px;
		clip-path: polygon(50% 0%, 95% 25%, 95% 75%, 50% 100%, 5% 75%, 5% 25%);
		display: flex;
		align-items: center;
		justify-content: center;
		transition: opacity 150ms ease;
		cursor: default;
	}

	.hex-node[data-rel="active"] {
		background: color-mix(in srgb, var(--accent) 30%, rgba(8,8,16,0.9));
		box-shadow: 0 0 16px color-mix(in srgb, var(--accent) 40%, transparent);
	}

	.hex-node[data-rel="past"] {
		background: rgba(255,255,255,0.04);
		opacity: 0.45;
	}

	.hex-node[data-rel="future"] {
		background: rgba(255,255,255,0.02);
		opacity: 0.25;
	}

	.hex-inner {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
	}

	.hex-name {
		font-size: 0.5625rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--hub-text, rgba(255,255,255,0.88));
	}

	.hex-node[data-rel="past"] .hex-name,
	.hex-node[data-rel="future"] .hex-name {
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
	}

	.active-dot {
		width: 5px;
		height: 5px;
		border-radius: 50%;
		background: var(--accent);
		animation: dot-pulse 1.8s ease-in-out infinite;
	}

	@keyframes dot-pulse {
		0%, 100% { opacity: 0.6; transform: scale(1); }
		50%       { opacity: 1.0; transform: scale(1.4); }
	}

	/* ── Active slot detail ── */
	.active-detail {
		display: flex;
		align-items: baseline;
		gap: 6px;
		font-size: var(--hub-text-xs, 0.6875rem);
		border-top: 1px solid var(--hub-border-subtle, rgba(255,255,255,0.05));
		padding-top: var(--hub-space-2, 8px);
		min-width: 0;
	}

	.active-who {
		color: var(--accent);
		font-weight: 600;
		letter-spacing: 0.04em;
		flex-shrink: 0;
	}

	.active-sep { color: var(--hub-text-subtle, rgba(255,255,255,0.28)); flex-shrink: 0; }

	.active-task {
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		flex: 1;
		min-width: 0;
	}

	.active-elapsed {
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		font-size: var(--hub-text-2xs, 0.625rem);
		flex-shrink: 0;
	}

	/* ── Misc ── */
	.loading, .err {
		font-size: var(--hub-text-xs, 0.6875rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		text-align: center;
		padding: 16px 0;
	}

	.err { color: var(--state-blocked, #f59e0b); }
</style>
