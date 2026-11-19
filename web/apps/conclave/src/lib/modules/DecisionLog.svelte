<script lang="ts">
	import { onMount } from 'svelte';
	import { HUB_URL } from '../store.svelte.js';

	/*
	 * DecisionLog — Wren / Mox.Design 2026-05-07
	 *
	 * Append-only decisions record — the flight data recorder display.
	 * Parses /mnt/KB/claude/_decisions/symmate-build.md via Keel's API.
	 *
	 * Pattern: ATC scope history trail (authoritative, uneditable, timestamped).
	 * No delete. No edit. Append-only is a design signal.
	 *
	 * Data: GET /api/decisions
	 */

	interface Decision {
		date: string;       // ISO date
		author: string;     // "Wren", "Elder", etc.
		title: string;      // first non-empty heading line
		body: string;       // full markdown body
		line_number: number;
	}

	// Author accent colors — map to sibling identity
	const AUTHOR_COLOR: Record<string, string> = {
		elder:  '#e8a878',
		keel:   '#4a9eff',
		quill:  '#8b5cf6',
		trace:  '#f59e0b',
		wren:   '#10b981',
		wright: '#3d8ef0',
		forge:  '#ec4899',
		loke:   'rgba(255,255,255,0.7)',
	};

	function authorColor(author: string): string {
		return AUTHOR_COLOR[author.toLowerCase()] ?? 'var(--hub-text-muted)';
	}

	let decisions = $state<Decision[]>([]);
	let expanded = $state<Set<number>>(new Set());
	let error = $state(false);

	function toggle(lineNo: number) {
		const next = new Set(expanded);
		if (next.has(lineNo)) next.delete(lineNo);
		else next.add(lineNo);
		expanded = next;
	}

	function formatDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString([], { month: 'short', day: 'numeric', year: 'numeric' });
		} catch { return iso; }
	}

	onMount(() => {
		async function load() {
			try {
				const r = await fetch(`${HUB_URL}/decisions`);
				if (!r.ok) throw new Error(`${r.status}`);
				const d = await r.json();
				decisions = (d.decisions ?? []).slice().reverse(); // newest first
				error = false;
			} catch {
				error = true;
			}
		}
		load();
		// Decisions are append-only, slow-changing — refresh every 30s
		const iv = setInterval(load, 30000);
		return () => clearInterval(iv);
	});
</script>

<div class="log">
	{#if error}
		<div class="error-state">
			⚠ /api/decisions unreachable — Keel endpoint needed
		</div>
	{:else if decisions.length === 0}
		<div class="empty-state">
			No decisions logged yet.<br>
			Append to <code>/mnt/KB/claude/_decisions/symmate-build.md</code>.
		</div>
	{:else}
		<div class="timeline">
			{#each decisions as d (d.line_number)}
				{@const open = expanded.has(d.line_number)}
				{@const color = authorColor(d.author)}

				<div class="entry">
					<div class="rail">
						<div class="rail-dot" style="background: {color}; box-shadow: 0 0 6px {color}33"></div>
						<div class="rail-line"></div>
					</div>

					<div class="content">
						<div class="meta">
							<span class="date">{formatDate(d.date)}</span>
							<span class="author" style="color: {color}">{d.author}</span>
						</div>
						<div class="title">{d.title}</div>

						{#if d.body && d.body.trim()}
							<button
								class="expand-btn"
								onclick={() => toggle(d.line_number)}
								aria-expanded={open}
							>
								{open ? '▲ collapse' : '▼ expand'}
							</button>

							{#if open}
								<pre class="body">{d.body.trim()}</pre>
							{/if}
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.log {
		height: 100%;
		overflow-y: auto;
		padding: var(--hub-space-3, 12px) var(--hub-space-3, 12px) var(--hub-space-3, 12px) var(--hub-space-2, 8px);
	}

	.error-state, .empty-state {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-xs, 0.6875rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		text-align: center;
		padding: 32px 16px;
		line-height: 1.7;
	}

	.error-state { color: var(--state-blocked, #f59e0b); }

	code {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		background: rgba(255,255,255,0.06);
		padding: 1px 4px;
		border-radius: 3px;
		font-size: inherit;
	}

	/* ── Timeline ── */
	.timeline {
		display: flex;
		flex-direction: column;
	}

	.entry {
		display: flex;
		gap: 10px;
		min-height: 48px;
	}

	/* ── Rail ── */
	.rail {
		display: flex;
		flex-direction: column;
		align-items: center;
		flex-shrink: 0;
		width: 14px;
	}

	.rail-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
		margin-top: 6px;
	}

	.rail-line {
		flex: 1;
		width: 1px;
		background: var(--hub-border, rgba(255,255,255,0.08));
		margin-top: 4px;
	}

	.entry:last-child .rail-line { display: none; }

	/* ── Content ── */
	.content {
		flex: 1;
		min-width: 0;
		padding-bottom: var(--hub-space-4, 16px);
	}

	.meta {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 3px;
	}

	.date {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		letter-spacing: 0.05em;
	}

	.author {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	.title {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-xs, 0.6875rem);
		color: var(--hub-text, rgba(255,255,255,0.88));
		font-weight: 500;
		line-height: 1.4;
		margin-bottom: 4px;
	}

	.expand-btn {
		background: none;
		border: none;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		padding: 0;
		cursor: pointer;
		letter-spacing: 0.04em;
		transition: color 120ms ease;
	}
	.expand-btn:hover { color: var(--hub-text-muted, rgba(255,255,255,0.5)); }

	.body {
		margin-top: 8px;
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
		white-space: pre-wrap;
		word-break: break-word;
		line-height: 1.65;
		background: rgba(0, 0, 0, 0.25);
		border-left: 2px solid var(--hub-border, rgba(255,255,255,0.08));
		padding: 8px 10px;
		border-radius: 0 4px 4px 0;
	}
</style>
