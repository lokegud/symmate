<script lang="ts">
	import { onMount } from 'svelte';
	import { HUB_URL } from '../store.svelte.js';

	/*
	 * ResourceMeter — Wren / Mox.Design 2026-05-07
	 *
	 * Parlour + stak hardware state at a glance.
	 * Pattern: ship bridge conning display — primary values always readable,
	 * warning colors threshold-triggered (>80% amber, >95% red).
	 *
	 * Data: GET /resources (Keel, polls every 10s)
	 * Shape: { fetched_at: string, hosts: HostMetrics[] }
	 */

	interface GpuMetrics {
		name: string;
		util_pct: number;
		vram_used_gb: number;
		vram_total_gb: number;
		active_process: string | null;
	}

	interface HostMetrics {
		host: string;
		cpu_pct: number;
		ram_used_gb: number;
		ram_total_gb: number;
		ram_pct: number;
		gpu?: GpuMetrics;
	}

	interface ResourcesResponse {
		fetched_at: string;
		hosts: HostMetrics[];
	}

	let data = $state<ResourcesResponse | null>(null);
	let error = $state(false);
	let fetchedAt = $state('');

	// Color threshold logic — matches bindrune-v2 state semantics
	function barColor(pct: number): string {
		if (pct >= 95) return 'var(--state-blocked-ring, #cc4422)';
		if (pct >= 80) return 'var(--state-blocked, #f59e0b)';
		return 'var(--hub-accent, #4a9eff)';
	}

	function labelColor(pct: number): string {
		if (pct >= 95) return 'var(--state-blocked-ring, #cc4422)';
		if (pct >= 80) return 'var(--state-blocked, #f59e0b)';
		return 'var(--hub-text-muted, rgba(255,255,255,0.5))';
	}

	function formatGb(n: number): string {
		return n.toFixed(1);
	}

	function formatPct(n: number): string {
		return n.toFixed(1) + '%';
	}

	function formatTime(iso: string): string {
		try { return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }); }
		catch { return ''; }
	}

	onMount(() => {
		async function poll() {
			try {
				const r = await fetch(`${HUB_URL}/resources`);
				if (!r.ok) throw new Error(`${r.status}`);
				const d: ResourcesResponse = await r.json();
				data = d;
				fetchedAt = formatTime(d.fetched_at);
				error = false;
			} catch {
				error = true;
			}
		}
		poll();
		const iv = setInterval(poll, 10000);
		return () => clearInterval(iv);
	});
</script>

<div class="meter">
	{#if error}
		<div class="empty">⚠ /resources unreachable</div>
	{:else if !data}
		<!-- Skeleton -->
		{#each Array(2) as _}
			<div class="host-skeleton">
				<div class="skel-label"></div>
				<div class="skel-bar"></div>
				<div class="skel-bar short"></div>
			</div>
		{/each}
	{:else}
		{#each data.hosts as host (host.host)}
			<div class="host-block">
				<div class="host-name">{host.host}</div>

				<!-- CPU -->
				<div class="metric-row">
					<span class="metric-label">CPU</span>
					<div class="bar-track">
						<div
							class="bar-fill"
							style="width: {Math.min(host.cpu_pct, 100)}%; background: {barColor(host.cpu_pct)}"
						></div>
					</div>
					<span class="metric-val" style="color: {labelColor(host.cpu_pct)}">
						{formatPct(host.cpu_pct)}
					</span>
				</div>

				<!-- RAM -->
				<div class="metric-row">
					<span class="metric-label">RAM</span>
					<div class="bar-track">
						<div
							class="bar-fill"
							style="width: {Math.min(host.ram_pct, 100)}%; background: {barColor(host.ram_pct)}"
						></div>
					</div>
					<span class="metric-val" style="color: {labelColor(host.ram_pct)}">
						{formatGb(host.ram_used_gb)} / {formatGb(host.ram_total_gb)} GB
					</span>
				</div>

				<!-- GPU (optional) -->
				{#if host.gpu}
					{@const gpu = host.gpu}
					<div class="metric-row">
						<span class="metric-label">GPU</span>
						<div class="bar-track">
							<div
								class="bar-fill"
								style="width: {Math.min(gpu.util_pct, 100)}%; background: {barColor(gpu.util_pct)}"
							></div>
						</div>
						<span class="metric-val" style="color: {labelColor(gpu.util_pct)}">
							{formatPct(gpu.util_pct)}
						</span>
					</div>
					<div class="gpu-detail">
						<span class="metric-label">VRAM</span>
						<span class="gpu-name">{formatGb(gpu.vram_used_gb)} / {formatGb(gpu.vram_total_gb)} GB</span>
						{#if gpu.active_process}
							<span class="gpu-proc">· {gpu.active_process}</span>
						{/if}
					</div>
				{:else}
					<div class="metric-row muted">
						<span class="metric-label">GPU</span>
						<span class="metric-val muted">—</span>
					</div>
				{/if}
			</div>
		{/each}

		{#if fetchedAt}
			<div class="footer">updated {fetchedAt}</div>
		{/if}
	{/if}
</div>

<style>
	.meter {
		display: flex;
		flex-direction: column;
		gap: var(--hub-space-3, 12px);
		padding: var(--hub-space-3, 12px);
		height: 100%;
		overflow-y: auto;
		font-family: var(--hub-font-mono, ui-monospace, monospace);
	}

	.empty {
		font-size: var(--hub-text-xs, 0.6875rem);
		color: var(--state-blocked, #f59e0b);
		text-align: center;
		padding: 32px 0;
		letter-spacing: 0.05em;
	}

	/* ── Host block ── */
	.host-block {
		display: flex;
		flex-direction: column;
		gap: 5px;
		padding-bottom: var(--hub-space-3, 12px);
		border-bottom: 1px solid var(--hub-border-subtle, rgba(255,255,255,0.05));
	}

	.host-block:last-of-type { border-bottom: none; }

	.host-name {
		font-size: var(--hub-text-2xs, 0.625rem);
		font-weight: 600;
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		letter-spacing: 0.1em;
		text-transform: uppercase;
		margin-bottom: 3px;
	}

	/* ── Metric row ── */
	.metric-row {
		display: grid;
		grid-template-columns: 36px 1fr 90px;
		align-items: center;
		gap: 8px;
	}

	.metric-label {
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		letter-spacing: 0.06em;
		text-align: right;
	}

	.metric-val {
		font-size: var(--hub-text-2xs, 0.625rem);
		text-align: right;
		letter-spacing: 0.03em;
		white-space: nowrap;
	}

	.metric-row.muted .metric-val { color: var(--hub-text-subtle, rgba(255,255,255,0.28)); }

	/* ── Bar ── */
	.bar-track {
		height: 4px;
		background: rgba(255,255,255,0.06);
		border-radius: 2px;
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		border-radius: 2px;
		transition: width 400ms var(--hub-ease, ease), background 400ms ease;
	}

	/* ── GPU detail sub-row ── */
	.gpu-detail {
		display: flex;
		align-items: center;
		gap: 6px;
		padding-left: 44px; /* aligns under the bar */
	}

	.gpu-name {
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-muted, rgba(255,255,255,0.5));
	}

	.gpu-proc {
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-accent, #4a9eff);
		opacity: 0.7;
	}

	/* ── Footer ── */
	.footer {
		font-size: var(--hub-text-2xs, 0.625rem);
		color: var(--hub-text-subtle, rgba(255,255,255,0.28));
		text-align: right;
		letter-spacing: 0.04em;
		margin-top: auto;
	}

	/* ── Skeleton ── */
	.host-skeleton {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding-bottom: 12px;
	}

	.skel-label, .skel-bar {
		background: rgba(255,255,255,0.06);
		border-radius: 2px;
		animation: skel-pulse 1.2s ease-in-out infinite alternate;
	}

	.skel-label { height: 8px; width: 40px; }
	.skel-bar   { height: 4px; width: 100%; }
	.skel-bar.short { width: 60%; }

	@keyframes skel-pulse {
		from { opacity: 0.3; }
		to   { opacity: 0.6; }
	}
</style>
