<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchConclaveMessages, type ConclaveMessage } from '../store.svelte.js';

	/*
	 * ConversationStream — Wren / Mox.Design 2026-05-07
	 *
	 * Renders the Matrix conclave thread live.
	 * Pattern: ham radio JS8Call decoded log + ATC scope history trail.
	 *
	 * - New messages at bottom, old scrolls up
	 * - "Follow" auto-scroll, pause on manual scroll
	 * - PAUSE messages: full-width amber highlight
	 * - Elder round-robin messages: blue-accent left border
	 * - Polled every 5s; replace with SSE when Keel ships the stream
	 */

	let messages = $state<ConclaveMessage[]>([]);
	let following = $state(true);
	let newCount = $state(0);
	let error = $state(false);
	let scrollEl: HTMLElement;
	let bottomEl: HTMLElement;

	function shortSender(sender: string): string {
		return sender.replace(/^@/, '').split(':')[0];
	}

	function formatTime(iso: string): string {
		try { return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }); }
		catch { return '??:??'; }
	}

	function isPause(body: string): boolean {
		return /^\s*PAUSE\b/i.test(body);
	}

	function isRoundRobin(body: string, sender: string): boolean {
		return /^@?elder/.test(sender) && /\*\*.*cohort|round|task \d|change [A-Z]/i.test(body);
	}

	function isLoke(sender: string): boolean {
		return /^@?loke/.test(sender);
	}

	function scrollToBottom(smooth = true) {
		bottomEl?.scrollIntoView({ behavior: smooth ? 'smooth' : 'instant', block: 'end' });
	}

	function onScroll() {
		if (!scrollEl) return;
		const atBottom = scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight < 40;
		if (atBottom) { following = true; newCount = 0; }
		else if (following) { following = false; }
	}

	function followClick() {
		following = true;
		newCount = 0;
		scrollToBottom();
	}

	onMount(() => {
		async function poll() {
			try {
				const data = await fetchConclaveMessages(60);
				const prevLen = messages.length;
				messages = data;
				error = false;
				if (following) {
					// tiny delay to let DOM update
					setTimeout(() => scrollToBottom(prevLen > 0), 10);
				} else {
					const added = data.length - prevLen;
					if (added > 0) newCount += added;
				}
			} catch {
				error = true;
			}
		}
		poll();
		const iv = setInterval(poll, 5000);
		return () => clearInterval(iv);
	});
</script>

<div class="stream">
	{#if error}
		<div class="error-banner">⚠ conclave/messages unreachable</div>
	{/if}

	<div class="messages" bind:this={scrollEl} onscroll={onScroll}>
		{#if messages.length === 0 && !error}
			<!-- Skeleton loading rows -->
			{#each Array(6) as _, i}
				<div class="skeleton-row" style="opacity: {0.3 + (i % 2) * 0.1}"></div>
			{/each}
		{:else}
			{#each messages as msg (msg.ts + msg.sender)}
				{@const sender = shortSender(msg.sender)}
				{@const pause = isPause(msg.body)}
				{@const rr    = !pause && isRoundRobin(msg.body, sender)}
				{@const loke  = !pause && !rr && isLoke(msg.sender)}
				<div
					class="row"
					class:is-pause={pause}
					class:is-rr={rr}
					class:is-loke={loke}
				>
					<span class="ts">{formatTime(msg.ts)}</span>
					<span class="mark" title={msg.sender}>
						{sender.slice(0, 2).toUpperCase()}
					</span>
					<span class="sender">{sender}</span>
					<span class="body">{msg.body}</span>
				</div>
			{/each}
		{/if}
		<div bind:this={bottomEl}></div>
	</div>

	{#if !following}
		<button class="follow-btn" onclick={followClick}>
			↓{newCount > 0 ? ` ${newCount} new` : ' Follow'}
		</button>
	{/if}
</div>

<style>
	.stream {
		display: flex;
		flex-direction: column;
		height: 100%;
		position: relative;
		overflow: hidden;
	}

	.error-banner {
		background: rgba(245, 158, 11, 0.1);
		border-bottom: 1px solid var(--state-blocked-ring, #cc4422);
		color: var(--state-blocked, #f59e0b);
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.625rem;
		padding: 4px 12px;
		letter-spacing: 0.05em;
		flex-shrink: 0;
	}

	.messages {
		flex: 1;
		overflow-y: auto;
		overflow-x: hidden;
		padding: 6px 0;
	}

	/* ── Message row ── */
	.row {
		display: grid;
		grid-template-columns: 48px 24px 72px 1fr;
		gap: 0 6px;
		padding: 3px 10px 3px 8px;
		border-left: 2px solid transparent;
		align-items: baseline;
		min-height: 22px;
	}

	.row:hover { background: rgba(255,255,255,0.02); }

	.row.is-pause {
		background: rgba(245, 158, 11, 0.07);
		border-left-color: var(--state-blocked, #f59e0b);
	}

	.row.is-rr {
		border-left-color: var(--hub-accent, #4a9eff);
	}

	.row.is-loke .sender {
		color: var(--hub-text, rgba(255,255,255,0.88));
	}

	.ts {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.625rem;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		letter-spacing: 0.03em;
		text-align: right;
		flex-shrink: 0;
	}

	.mark {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.5625rem;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		background: rgba(255,255,255,0.05);
		border-radius: 2px;
		width: 22px;
		height: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		letter-spacing: 0.02em;
		align-self: center;
		flex-shrink: 0;
	}

	.sender {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.6875rem;
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.body {
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.6875rem;
		color: var(--hub-text, rgba(255,255,255,0.88));
		white-space: pre-wrap;
		word-break: break-word;
		line-height: 1.45;
		min-width: 0;
	}

	.row.is-pause .body { color: var(--state-blocked, #f59e0b); font-weight: 600; }

	/* ── Skeleton ── */
	.skeleton-row {
		height: 20px;
		margin: 3px 10px;
		border-radius: 3px;
		background: rgba(255,255,255,0.06);
		animation: skel-pulse 1.2s ease-in-out infinite alternate;
	}

	@keyframes skel-pulse {
		from { opacity: 0.3; }
		to   { opacity: 0.6; }
	}

	/* ── Follow button ── */
	.follow-btn {
		position: absolute;
		bottom: 10px;
		right: 10px;
		background: var(--hub-accent-dim, rgba(74,158,255,0.15));
		border: 1px solid var(--hub-accent-border, rgba(74,158,255,0.35));
		border-radius: var(--hub-radius-full, 9999px);
		color: var(--hub-accent, #4a9eff);
		font-family: var(--hub-font-mono, ui-monospace, monospace);
		font-size: 0.625rem;
		padding: 4px 12px;
		cursor: pointer;
		letter-spacing: 0.05em;
		transition: background 120ms ease;
		backdrop-filter: blur(8px);
	}
	.follow-btn:hover {
		background: rgba(74, 158, 255, 0.25);
	}
</style>
