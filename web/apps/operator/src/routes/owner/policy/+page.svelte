<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let toolTags = $state<string>((data.policy.tool_tags ?? []).join('\n'));

	$effect(() => { toolTags = (data.policy.tool_tags ?? []).join('\n'); });
	let saving = $state(false);

	// Must match the tags the runtime actually registers (internal/tool).
	const KNOWN_TAGS = [
		{ tag: 'time', desc: 'Date and time lookup' },
		{ tag: 'net', desc: 'Web fetch and search' },
		{ tag: 'engineering', desc: 'GitHub: PRs, issues, deploy risk, cycle time' }
	];

	function toggleTag(tag: string) {
		const lines = toolTags.split('\n').map((t) => t.trim()).filter(Boolean);
		const idx = lines.indexOf(tag);
		if (idx === -1) {
			lines.push(tag);
		} else {
			lines.splice(idx, 1);
		}
		toolTags = lines.join('\n');
	}

	function isActive(tag: string) {
		return toolTags.split('\n').map((t) => t.trim()).includes(tag);
	}
</script>

<svelte:head>
	<title>Policy · Owner Dashboard</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1 class="font-display">Policy</h1>
		<p class="subtitle">Control which tools your Symmate can use. An empty list means all registered tools are available.</p>
	</header>

	{#if form?.success}
		<div class="banner banner--success">Policy saved.</div>
	{/if}
	{#if form?.error}
		<div class="banner banner--error">{form.error}</div>
	{/if}

	<form
		method="POST"
		use:enhance={() => {
			saving = true;
			return async ({ update }) => { saving = false; update(); };
		}}
	>
		<section class="section">
			<div class="section-head">
				<h2>Tool grants</h2>
				<p>Toggle known tool categories or enter custom tag prefixes directly.</p>
			</div>

			<div class="tag-grid">
				{#each KNOWN_TAGS as { tag, desc }}
					<button
						type="button"
						class="tag-btn"
						class:tag-btn--active={isActive(tag)}
						onclick={() => toggleTag(tag)}
						aria-pressed={isActive(tag)}
					>
						<span class="tag-name">{tag}</span>
						<span class="tag-desc">{desc}</span>
					</button>
				{/each}
			</div>

			<div class="field">
				<label for="tool_tags">Raw tool tags <span class="optional">(one per line)</span></label>
				<textarea
					id="tool_tags"
					name="tool_tags"
					bind:value={toolTags}
					rows="6"
					placeholder="web&#10;datetime&#10;calendar"
					disabled={saving}
				></textarea>
				<p class="hint">Prefix match — "net" allows all tools tagged "net" or "net/…". Empty = allow all. This list limits everyone, including you; who may use restricted tools is set on the Users page.</p>
			</div>

			<div class="firewall-notice">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
				<span>The privacy firewall runs independently of tool grants and cannot be disabled here.</span>
			</div>

			<div class="form-footer">
				<button type="submit" class="btn btn--primary" disabled={saving}>
					{saving ? 'Saving…' : 'Save policy'}
				</button>
			</div>
		</section>
	</form>
</div>

<style>
	.page {
		max-width: 700px;
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

	.banner {
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-md);
		font-size: var(--text-sm);
	}

	.banner--success { background: var(--color-success-subtle); color: var(--color-success); }
	.banner--error   { background: var(--color-error-subtle);   color: var(--color-error); }

	.section {
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-lg);
		padding: var(--space-8);
		display: flex;
		flex-direction: column;
		gap: var(--space-6);
		box-shadow: var(--shadow-sm);
	}

	.section-head h2 {
		font-size: var(--text-xl);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}

	.section-head p {
		margin-top: var(--space-1);
		font-size: var(--text-sm);
		color: var(--color-text-muted);
	}

	.tag-grid {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.tag-btn {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 2px;
		padding: var(--space-3) var(--space-4);
		background: var(--color-surface-raised);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-md);
		cursor: pointer;
		font-family: var(--font-body);
		text-align: left;
		transition: border-color 100ms, background-color 100ms;
	}

	.tag-btn:hover {
		border-color: var(--color-border-strong);
		background: var(--color-surface);
	}

	.tag-btn--active {
		border-color: var(--color-accent);
		background: var(--color-accent-subtle);
	}

	.tag-name {
		font-size: var(--text-sm);
		font-weight: var(--weight-semibold);
		font-family: var(--font-mono);
		color: var(--color-text);
	}

	.tag-btn--active .tag-name { color: var(--color-accent); }

	.tag-desc {
		font-size: var(--text-xs);
		color: var(--color-text-muted);
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1-5);
	}

	label {
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		color: var(--color-text);
	}

	.optional {
		color: var(--color-text-subtle);
		font-weight: var(--weight-regular);
	}

	textarea {
		background: var(--color-surface-raised);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-md);
		padding: var(--space-3) var(--space-4);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		line-height: 1.6;
		color: var(--color-text);
		outline: none;
		resize: vertical;
		width: 100%;
		transition: border-color 100ms, box-shadow 100ms;
	}

	textarea:focus {
		border-color: var(--color-accent);
		box-shadow: 0 0 0 3px var(--color-accent-subtle);
	}

	textarea:disabled { opacity: 0.5; }

	.hint {
		font-size: var(--text-xs);
		color: var(--color-text-subtle);
	}

	.firewall-notice {
		display: flex;
		align-items: flex-start;
		gap: var(--space-2);
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		background: var(--color-surface-raised);
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-md);
	}

	.firewall-notice svg { flex-shrink: 0; margin-top: 2px; }

	.form-footer {
		display: flex;
		justify-content: flex-end;
	}

	.btn {
		height: 40px;
		padding: 0 var(--space-6);
		border: none;
		border-radius: var(--radius-md);
		font-family: var(--font-body);
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		cursor: pointer;
		transition: background-color 100ms;
	}

	.btn--primary {
		background: var(--color-accent);
		color: var(--color-accent-text);
	}

	.btn--primary:hover:not(:disabled) { background: var(--color-accent-hover); }
	.btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
