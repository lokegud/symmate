<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let contextVal = $state<string>(data.personality.context);
	let verticalVal = $state<string>(data.personality.vertical);
	let provider = $state<string>(data.llmConfig?.provider ?? 'hermes');
	let model = $state<string>(data.llmConfig?.model ?? '');
	let customEndpoint = $state<string>(data.llmConfig?.custom_endpoint ?? '');
	let showCustomKey = $state(false);

	// Sync initial values if data changes (e.g. after form action).
	$effect(() => {
		contextVal = data.personality.context;
		verticalVal = data.personality.vertical;
		provider = data.llmConfig?.provider ?? 'hermes';
		model = data.llmConfig?.model ?? '';
		customEndpoint = data.llmConfig?.custom_endpoint ?? '';
	});

	let savingPersonality = $state(false);
	let savingLLM = $state(false);

	const PROVIDERS = [
		{ value: 'anthropic', label: 'Anthropic' },
		{ value: 'hermes', label: 'Hermes (local)' },
		{ value: 'openai-compat', label: 'OpenAI-compatible' }
	];

	const ANTHROPIC_MODELS = [
		{ value: 'claude-sonnet-4-6', label: 'Sonnet 4.6 (default)' },
		{ value: 'claude-haiku-4-5', label: 'Haiku 4.5 (fast)' },
		{ value: 'claude-opus-4-7', label: 'Opus 4.7 (best)' }
	];
</script>

<svelte:head>
	<title>Personality · Owner Dashboard</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1 class="font-display">Personality</h1>
		<p class="subtitle">Edit how your Symmate presents itself. Changes take effect on the next conversation.</p>
	</header>

	<!-- Personality Editor -->
	<section class="section">
		<div class="section-head">
			<h2>Context</h2>
			<p>Defines who your Symmate is — name, background, communication style.</p>
		</div>

		{#if form?.success}
			<div class="banner banner--success">Personality saved.</div>
		{/if}
		{#if form?.error}
			<div class="banner banner--error">{form.error}</div>
		{/if}

		<form
			method="POST"
			action="?/savePersonality"
			use:enhance={() => {
				savingPersonality = true;
				return async ({ update }) => {
					savingPersonality = false;
					update();
				};
			}}
		>
			<div class="field">
				<label for="context">context.md</label>
				<textarea
					id="context"
					name="context"
					bind:value={contextVal}
					rows="12"
					placeholder="# Your name&#10;&#10;Describe your Symmate's context, personality, and purpose here…"
					disabled={savingPersonality}
				></textarea>
				<p class="hint">Markdown. The core identity layer — always included in every conversation.</p>
			</div>

			<div class="field">
				<label for="vertical">vertical.md <span class="optional">(optional)</span></label>
				<textarea
					id="vertical"
					name="vertical"
					bind:value={verticalVal}
					rows="8"
					placeholder="Domain-specific behaviour overrides. Leave blank for the default vertical."
					disabled={savingPersonality}
				></textarea>
				<p class="hint">Markdown. Overrides for a specific domain or business type.</p>
			</div>

			<div class="form-footer">
				<button type="submit" class="btn btn--primary" disabled={savingPersonality}>
					{savingPersonality ? 'Saving…' : 'Save personality'}
				</button>
			</div>
		</form>
	</section>

	<!-- Model / Provider Selector -->
	<section class="section">
		<div class="section-head">
			<h2>Model & Provider</h2>
			<p>Switch the LLM powering your Symmate without restarting it. Privacy-bound requests always use local Hermes.</p>
		</div>

		{#if form?.llmSuccess}
			<div class="banner banner--success">Model saved.</div>
		{/if}
		{#if form?.llmError}
			<div class="banner banner--error">{form.llmError}</div>
		{/if}

		<form
			method="POST"
			action="?/saveLLM"
			use:enhance={() => {
				savingLLM = true;
				return async ({ update }) => {
					savingLLM = false;
					update();
				};
			}}
		>
			<div class="field-row">
				<div class="field">
					<label for="provider">Provider</label>
					<select id="provider" name="provider" bind:value={provider} disabled={savingLLM}>
						{#each PROVIDERS as p}
							<option value={p.value}>{p.label}</option>
						{/each}
					</select>
				</div>

				{#if provider === 'anthropic'}
					<div class="field">
						<label for="model">Model</label>
						<select id="model" name="model" bind:value={model} disabled={savingLLM}>
							{#each ANTHROPIC_MODELS as m}
								<option value={m.value}>{m.label}</option>
							{/each}
						</select>
					</div>
				{:else}
					<div class="field">
						<label for="model">Model name</label>
						<input
							id="model"
							name="model"
							type="text"
							bind:value={model}
							placeholder={provider === 'hermes' ? 'hermes' : 'model-name'}
							disabled={savingLLM}
						/>
					</div>
				{/if}
			</div>

			{#if provider === 'openai-compat'}
				<div class="field">
					<label for="custom_endpoint">Endpoint URL</label>
					<input
						id="custom_endpoint"
						name="custom_endpoint"
						type="url"
						bind:value={customEndpoint}
						placeholder="https://api.example.com/v1"
						disabled={savingLLM}
					/>
				</div>

				<div class="field">
					<div class="label-row">
						<label for="custom_api_key">API Key</label>
						{#if data.llmConfig?.has_custom_key}
							<span class="key-set">Key stored · <button type="button" class="link-btn" onclick={() => (showCustomKey = !showCustomKey)}>{showCustomKey ? 'hide' : 'change'}</button></span>
						{/if}
					</div>
					{#if showCustomKey || !data.llmConfig?.has_custom_key}
						<input
							id="custom_api_key"
							name="custom_api_key"
							type="password"
							placeholder="sk-…"
							autocomplete="new-password"
							disabled={savingLLM}
						/>
					{:else}
						<input type="hidden" name="custom_api_key" value="" />
					{/if}
				</div>
			{/if}

			<div class="provider-notice">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
				Privacy-bound requests (email, messages, personal data) always route to local Hermes, ignoring this setting.
			</div>

			<div class="form-footer">
				<button type="submit" class="btn btn--primary" disabled={savingLLM}>
					{savingLLM ? 'Saving…' : 'Save model'}
				</button>
			</div>
		</form>
	</section>
</div>

<style>
	.page {
		max-width: 800px;
		display: flex;
		flex-direction: column;
		gap: var(--space-12);
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

	.banner {
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-md);
		font-size: var(--text-sm);
	}

	.banner--success { background: var(--color-success-subtle); color: var(--color-success); }
	.banner--error   { background: var(--color-error-subtle);   color: var(--color-error); }

	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1-5);
	}

	.field-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4);
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

	textarea, input[type="text"], input[type="url"], input[type="password"], select {
		background: var(--color-surface-raised);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-md);
		font-family: var(--font-body);
		font-size: var(--text-base);
		color: var(--color-text);
		outline: none;
		transition: border-color 100ms, box-shadow 100ms;
		width: 100%;
	}

	textarea {
		padding: var(--space-3) var(--space-4);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		line-height: 1.6;
		resize: vertical;
	}

	input[type="text"],
	input[type="url"],
	input[type="password"],
	select {
		height: 40px;
		padding: 0 var(--space-4);
	}

	select { cursor: pointer; }

	textarea:focus, input:focus, select:focus {
		border-color: var(--color-accent);
		box-shadow: 0 0 0 3px var(--color-accent-subtle);
	}

	textarea:disabled, input:disabled, select:disabled { opacity: 0.5; }

	.hint {
		font-size: var(--text-xs);
		color: var(--color-text-subtle);
	}

	.label-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.key-set {
		font-size: var(--text-xs);
		color: var(--color-text-muted);
	}

	.link-btn {
		background: none;
		border: none;
		padding: 0;
		font-size: inherit;
		color: var(--color-accent);
		cursor: pointer;
		font-family: inherit;
	}

	.provider-notice {
		display: flex;
		align-items: flex-start;
		gap: var(--space-2);
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		background: var(--color-surface-raised);
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-md);
	}

	.provider-notice svg {
		flex-shrink: 0;
		margin-top: 2px;
	}

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

	@media (max-width: 640px) {
		.field-row { grid-template-columns: 1fr; }
	}
</style>
