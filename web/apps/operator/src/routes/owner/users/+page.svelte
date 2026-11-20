<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let busy = $state(false);
	let copied = $state(false);

	const groups = $derived(data.users.available_groups);

	function submitting() {
		busy = true;
		copied = false;
		return async ({ update }: { update: (opts?: { reset?: boolean }) => Promise<void> }) => {
			busy = false;
			await update({ reset: false });
		};
	}

	async function copyToken(token: string) {
		try {
			await navigator.clipboard.writeText(token);
			copied = true;
		} catch {
			copied = false;
		}
	}

	function formatDate(s?: string) {
		if (!s) return 'never';
		const d = new Date(s);
		return isNaN(d.getTime()) || d.getFullYear() < 2000 ? 'never' : d.toLocaleString();
	}
</script>

<svelte:head>
	<title>Users · Owner Dashboard</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1 class="font-display">Users</h1>
		<p class="subtitle">
			Restricted tools (GitHub and other integrations) are yours only by default. Register someone here
			and toggle on the tool groups they may use. Anonymous visitors only ever get the public tools.
		</p>
	</header>

	{#if form?.newToken}
		<div class="banner banner--token" role="status">
			<p><strong>Access token for {form.tokenFor}.</strong> Copy it now — it won't be shown again.</p>
			<div class="token-row">
				<code class="token">{form.newToken}</code>
				<button type="button" class="btn btn--small" onclick={() => copyToken(form!.newToken!)}>
					{copied ? 'Copied' : 'Copy'}
				</button>
			</div>
			<p class="hint">They send it as <code>user_token</code> in the chat request, or as the <code>X-Symmate-User</code> header.</p>
		</div>
	{:else if form?.success}
		<div class="banner banner--success" role="status">{form.success}</div>
	{/if}
	{#if form?.error}
		<div class="banner banner--error" role="alert">{form.error}</div>
	{/if}

	<!-- Master switch -->
	<section class="section">
		<form method="POST" action="?/settings" use:enhance={submitting} class="switch-row">
			<div>
				<h2>User access</h2>
				<p class="muted">
					{data.users.enabled
						? 'On — registered users get the tool groups you toggled for them.'
						: 'Off — everyone except you is treated as an anonymous visitor.'}
				</p>
			</div>
			<input type="hidden" name="enabled" value={data.users.enabled ? 'false' : 'true'} />
			<button
				type="submit"
				class="switch"
				class:switch--on={data.users.enabled}
				role="switch"
				aria-checked={data.users.enabled}
				aria-label="User access"
				disabled={busy}
			>
				<span class="switch__knob"></span>
			</button>
		</form>
	</section>

	<!-- Add user -->
	<section class="section">
		<div class="section-head">
			<h2>Add a user</h2>
			<p class="muted">A token is generated for them. You can change their groups any time.</p>
		</div>
		<form method="POST" action="?/create" use:enhance={submitting} class="stack">
			<div class="field">
				<label for="new-name">Name</label>
				<input id="new-name" name="name" type="text" maxlength="80" required placeholder="e.g. Bob" disabled={busy} />
			</div>
			{#if groups.length > 0}
				<fieldset class="groups">
					<legend>Tool groups</legend>
					{#each groups as g}
						<label class="chip">
							<input type="checkbox" name="groups" value={g} disabled={busy} />
							<span>{g}</span>
						</label>
					{/each}
				</fieldset>
			{:else}
				<p class="hint">No tool groups to grant: either no integrations are enabled (set GITHUB_TOKEN or GITHUB_REPO for GitHub) or the Policy allowlist blocks them all.</p>
			{/if}
			<div class="form-footer">
				<button type="submit" class="btn btn--primary" disabled={busy}>Add user</button>
			</div>
		</form>
	</section>

	<!-- Existing users -->
	<section class="section">
		<div class="section-head">
			<h2>Registered users</h2>
		</div>
		{#if data.users.users.length === 0}
			<p class="muted">Nobody yet.</p>
		{:else}
			<ul class="user-list">
				{#each data.users.users as u (u.id)}
					<li class="user" class:user--off={!u.enabled}>
						<form method="POST" action="?/update" use:enhance={submitting} class="stack">
							<input type="hidden" name="id" value={u.id} />
							<div class="user__head">
								<div>
									<span class="user__name">{u.name}</span>
									<span class="user__meta">Last seen {formatDate(u.last_seen)}</span>
								</div>
								<label class="inline-toggle">
									<input type="checkbox" name="enabled" value="true" checked={u.enabled} disabled={busy} />
									<span>Enabled</span>
								</label>
							</div>
							{#if groups.length > 0}
								<fieldset class="groups">
									<legend class="sr-only">Tool groups for {u.name}</legend>
									{#each groups as g}
										<label class="chip">
											<input type="checkbox" name="groups" value={g} checked={u.groups.includes(g)} disabled={busy} />
											<span>{g}</span>
										</label>
									{/each}
								</fieldset>
							{/if}
							<div class="user__actions">
								<button type="submit" class="btn btn--primary btn--small" disabled={busy}>Save</button>
								<button type="submit" formaction="?/rotate" class="btn btn--ghost btn--small" disabled={busy}>New token</button>
								<input type="hidden" name="name" value={u.name} />
								<button
									type="submit"
									formaction="?/remove"
									class="btn btn--danger btn--small"
									disabled={busy}
									onclick={(e) => {
										if (!confirm(`Remove ${u.name}? Their token stops working immediately.`)) e.preventDefault();
									}}
								>Remove</button>
							</div>
						</form>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
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

	.muted { font-size: var(--text-sm); color: var(--color-text-muted); }
	.hint { font-size: var(--text-xs); color: var(--color-text-subtle); }

	.banner {
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-md);
		font-size: var(--text-sm);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}

	.banner--success { background: var(--color-success-subtle); color: var(--color-success); }
	.banner--error   { background: var(--color-error-subtle);   color: var(--color-error); }
	.banner--token   { background: var(--color-accent-subtle);  color: var(--color-text); }

	.token-row { display: flex; gap: var(--space-2); align-items: center; flex-wrap: wrap; }

	.token {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		background: var(--color-surface);
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-sm);
		word-break: break-all;
		flex: 1;
		min-width: 0;
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

	.section h2 {
		font-size: var(--text-xl);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}

	.section-head p { margin-top: var(--space-1); }

	.stack { display: flex; flex-direction: column; gap: var(--space-4); }

	.switch-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-4);
	}

	.switch {
		flex-shrink: 0;
		width: 52px;
		height: 30px;
		border-radius: 999px;
		border: 1.5px solid var(--color-border);
		background: var(--color-surface-raised);
		position: relative;
		cursor: pointer;
		transition: background-color 120ms, border-color 120ms;
	}

	.switch__knob {
		position: absolute;
		top: 3px;
		left: 3px;
		width: 21px;
		height: 21px;
		border-radius: 50%;
		background: var(--color-text-muted);
		transition: transform 120ms, background-color 120ms;
	}

	.switch--on { background: var(--color-accent); border-color: var(--color-accent); }
	.switch--on .switch__knob { transform: translateX(22px); background: var(--color-accent-text); }
	.switch:disabled { opacity: 0.5; cursor: not-allowed; }

	.field { display: flex; flex-direction: column; gap: var(--space-1-5); }

	label { font-size: var(--text-sm); font-weight: var(--weight-medium); color: var(--color-text); }

	input[type='text'] {
		height: 44px;
		background: var(--color-surface-raised);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-md);
		padding: 0 var(--space-4);
		font-family: var(--font-body);
		font-size: var(--text-base);
		color: var(--color-text);
		outline: none;
	}

	input[type='text']:focus {
		border-color: var(--color-accent);
		box-shadow: 0 0 0 3px var(--color-accent-subtle);
	}

	.groups {
		border: none;
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.groups legend {
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin-bottom: var(--space-2);
		width: 100%;
	}

	.chip {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		min-height: 44px;
		padding: 0 var(--space-4);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-md);
		background: var(--color-surface-raised);
		font-family: var(--font-mono);
		cursor: pointer;
	}

	.chip:has(input:checked) {
		border-color: var(--color-accent);
		background: var(--color-accent-subtle);
		color: var(--color-accent);
	}

	.user-list { list-style: none; display: flex; flex-direction: column; gap: var(--space-4); }

	.user {
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-md);
		padding: var(--space-4);
	}

	.user--off { opacity: 0.7; }

	.user__head { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); flex-wrap: wrap; }
	.user__name { display: block; font-weight: var(--weight-semibold); color: var(--color-text); }
	.user__meta { display: block; font-size: var(--text-xs); color: var(--color-text-subtle); }

	.inline-toggle { display: inline-flex; align-items: center; gap: var(--space-2); min-height: 44px; cursor: pointer; }

	.user__actions { display: flex; gap: var(--space-2); flex-wrap: wrap; }

	.form-footer { display: flex; justify-content: flex-end; }

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

	.btn--small { height: 44px; padding: 0 var(--space-4); font-size: var(--text-sm); }
	.btn--primary { background: var(--color-accent); color: var(--color-accent-text); }
	.btn--primary:hover:not(:disabled) { background: var(--color-accent-hover); }
	.btn--ghost { background: var(--color-surface-raised); color: var(--color-text); border: 1.5px solid var(--color-border); }
	.btn--danger { background: var(--color-error-subtle); color: var(--color-error); }
	.btn:disabled { opacity: 0.5; cursor: not-allowed; }

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		border: 0;
	}
</style>
