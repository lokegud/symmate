<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchCohortState, agentStore, HUB_URL, type CohortMember } from '../store.svelte.js';

	/*
	 * SystemLog — Wren / Mox.Design 2026-05-07
	 *
	 * Per-agent activity log. Primary: SSE stream from /pane/{agentId}
	 * (live tmux pane tail). Fallback: synthesised activity log from
	 * /cohort/state polls, showing task + status transitions.
	 *
	 * Pattern: SDR# console log + ATC history trail.
	 * Monospace, dense, timestamped. Auto-follows; pauseable.
	 *
	 * Upgrade path: when Keel ships GET /pane/{agentId} as SSE,
	 * set paneAvailable=true and the component switches automatically.
	 */

	interface LogLine {
		ts: string;    // HH:MM:SS
		text: string;
		type: 'stdout' | 'status' | 'error' | 'info';
	}

	// Agents come from agentStore or cohort poll — whichever is populated
	let members = $state<CohortMember[]>([]);
	let selectedId = $state<string>('');
	let lines = $state<Record<string, LogLine[]>>({});
	let following = $state(true);
	let paneAvailable = $state(false); // flips true if SSE connects successfully
	let sseConn: EventSource | null = null;
	let scrollEl: HTMLElement;
	let bottomEl: HTMLElement;

	const currentLines = $derived(lines[selectedId] ?? []);

	function nowHMS(): string {
		return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
	}

	function appendLine(agentId: string, text: string, type: LogLine['type'] = 'stdout') {
		const entry: LogLine = { ts: nowHMS(), text, type };
		lines = { ...lines, [agentId]: [...(lines[agentId] ?? []).slice(-500), entry] };
	}

	function scrollToBottom() {
		bottomEl?.scrollIntoView({ behavior: 'smooth', block: 'end' });
	}

	function onScroll() {
		if (!scrollEl) return;
		const atBottom = scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight < 40;
		if (atBottom) following = true;
		else following = false;
	}

	// Try SSE pane stream for the selected agent
	function connectPane(agentId: string) {
		sseConn?.close();
		sseConn = null;
		paneAvailable = false;
		if (!agentId) return;

		const es = new EventSource(`${HUB_URL}/pane/${agentId}`);

		es.onopen = () => { paneAvailable = true; };

		es.onmessage = (e) => {
			appendLine(agentId, e.data, 'stdout');
			if (following) setTimeout(scrollToBottom, 10);
		};

		es.onerror = () => {
			es.close();
			sseConn = null;
			paneAvailable = false;
			// Don't retry — SSE not available yet, fallback poll handles it
		};

		sseConn = es;
	}

	// Fallback: synthesise log lines from cohort state changes
	let lastTaskByAgent: Record<string, string> = {};
	let lastStatusByAgent: Record<string, string> = {};

	function ingestCohortState(data: CohortMember[]) {
		for (const m of data) {
			const prevTask   = lastTaskByAgent[m.id];
			const prevStatus = lastStatusByAgent[m.id];

			if (prevTask === undefined) {
				// First seen
				appendLine(m.id, `registered — status: ${m.status}, task: ${m.task || 'none'}`, 'info');
			} else {
				if (m.task !== prevTask && m.task) {
					appendLine(m.id, `task → ${m.task}`, 'status');
				}
				if (m.status !== prevStatus) {
					appendLine(m.id, `status → ${m.status}`, 'status');
				}
			}
			lastTaskByAgent[m.id]   = m.task;
			lastStatusByAgent[m.id] = m.status;
		}
	}

	$effect(() => {
		// When selected agent changes, reconnect SSE
		if (selectedId) connectPane(selectedId);
	});

	$effect(() => {
		if (following && currentLines.length > 0) {
			setTimeout(scrollToBottom, 10);
		}
	});

	onMount(() => {
		async function poll() {
			try {
				const data = await fetchCohortState();
				members = data;
				if (!selectedId && data.length > 0) selectedId = data[0].id;
				if (!paneAvailable) ingestCohortState(data);
			} catch { /* silent — agentStore fallback below */ }
		}

		poll();
		const iv = setInterval(poll, 5000);

		return () => {
			clearInterval(iv);
			sseConn?.close();
		};
	});

	// Seed member list from agentStore if cohort poll hasn't landed yet
	const displayMembers = $derived(
		members.length > 0
			? members
			: agentStore.agents.map(a => ({
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

	const lineTypeClass: Record<LogLine['type'], string> = {
		stdout: '',
		status: 'line-status',
		error:  'line-error',
		info:   'line-info',
	};
</script>

<div class="syslog">
	<!-- Agent tabs -->
	<div class="tabs">
		{#each displayMembers as m (m.id)}
			<button
				class="tab"
				class:active={selectedId === m.id}
				onclick={() => selectedId = m.id}
			>
				{m.name}
			</button>
		{/each}
		{#if displayMembers.length === 0}
			<span class="no-agents">No agents registered</span>
		{/if}

		<div class="tab-spacer"></div>

		<span class="source-indicator" class:live={paneAvailable}>
			{paneAvailable ? '● live' : '○ polled'}
		</span>
	</div>

	<!-- Log output -->
	<div class="output" bind:this={scrollEl} onscroll={onScroll}>
		{#if currentLines.length === 0}
			<div class="empty-log">
				{selectedId
					? (paneAvailable ? 'Waiting for pane output…' : 'Watching for task activity…')
					: 'Select an agent above'}
			</div>
		{:else}
			{#each currentLines as line, i (i)}
				<div class="line {lineTypeClass[line.type]}">
					<span class="line-ts">{line.ts}</span>
					<span class="line-text">{line.text}</span>
				</div>
			{/each}
		{/if}
		<div bind:this={bottomEl}></div>
	</div>

	<!-- Follow toggle -->
	{#if !following}
		<button class="follow-btn" onclick={() => { following = true; scrollToBottom(); }}>
			↓ Follow
		</button>
	{/if}
</div>

<style>
	.syslog {
		display: flex;
		flex-direction: column;
		height: 100%;
		position: relative;
		overflow: hidden;
	}

	/* ── Tabs ── */
	.tabs {
		display: flex;
		align-items: center;
		gap: 2px;
		padding: 4px 8px;
		border-bottom: 1px solid var(--hub-border-subtle, rgba(255,255,255,0.05));
		flex-shrink: 0;
		overflow-x: auto;
	}

	.tab {
		background: none;
		border: 1px solid transparent;
		border-radius: var(--hub-radius-sm, 4px);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		padding: 3px 8px;
		cursor: pointer;
		letter-spacing: 0.04em;
		transition: all 120ms ease;
		white-space: nowrap;
	}

	.tab:hover { color: var(--hub-text-muted, rgba(255,255,255,0.5)); }

	.tab.active {
		background: rgba(255,255,255,0.06);
		border-color: var(--hub-border, rgba(255,255,255,0.08));
		color: var(--hub-text, rgba(255,255,255,0.88));
	}

	.tab-spacer { flex: 1; }

	.source-indicator {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		letter-spacing: 0.05em;
		white-space: nowrap;
	}

	.source-indicator.live { color: var(--state-working, #10b981); }

	.no-agents {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
	}

	/* ── Log output ── */
	.output {
		flex: 1;
		overflow-y: auto;
		overflow-x: hidden;
		padding: 4px 0;
		background: rgba(0,0,0,0.15);
	}

	.empty-log {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-xs, 0.6875rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		text-align: center;
		padding: 32px 0;
		letter-spacing: 0.05em;
	}

	/* ── Log line ── */
	.line {
		display: grid;
		grid-template-columns: 72px 1fr;
		gap: 6px;
		padding: 1px 10px;
		min-height: 18px;
		align-items: baseline;
	}

	.line:hover { background: rgba(255,255,255,0.02); }

	.line-ts {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.5625rem;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		letter-spacing: 0.03em;
		text-align: right;
		flex-shrink: 0;
		user-select: none;
	}

	.line-text {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
		white-space: pre-wrap;
		word-break: break-all;
		line-height: 1.5;
	}

	/* Line type variants */
	.line-status .line-text { color: var(--hub-accent, #4a9eff); }
	.line-error  .line-text { color: var(--state-blocked-ring, #cc4422); }
	.line-info   .line-text { color: var(--hub-text-subtle, rgba(255,255,255,0.28)); font-style: italic; }

	/* ── Follow button ── */
	.follow-btn {
		position: absolute;
		bottom: 8px;
		right: 8px;
		background: rgba(0,0,0,0.6);
		border: 1px solid var(--hub-border, rgba(255,255,255,0.08));
		border-radius: var(--hub-radius-full, 9999px);
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: var(--hub-text-2xs, 0.625rem);
		padding: 3px 10px;
		cursor: pointer;
		letter-spacing: 0.05em;
		backdrop-filter: blur(8px);
		transition: border-color 120ms ease;
	}
	.follow-btn:hover { border-color: var(--hub-accent, #4a9eff); color: var(--hub-accent, #4a9eff); }
</style>
