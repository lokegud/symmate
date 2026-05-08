<script lang="ts">
	import '@symmate/design-system/minimal.css';
	import { enhance } from '$app/forms';
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
	let loading = $state(false);
</script>

<svelte:head>
	<title>Owner Login · Symmate</title>
</svelte:head>

<div class="login-shell">
	<div class="login-card">
		<div class="login-header">
			<div class="blob" aria-hidden="true"></div>
			<h1 class="font-display">Owner Dashboard</h1>
			<p class="subtitle">Enter your recognition phrase to continue.</p>
		</div>

		{#if form?.error}
			<div class="error-banner" role="alert">{form.error}</div>
		{/if}

		<form
			method="POST"
			use:enhance={() => {
				loading = true;
				return async ({ update }) => {
					loading = false;
					update();
				};
			}}
		>
			<div class="field">
				<label for="phrase">Phrase</label>
				<input
					id="phrase"
					name="phrase"
					type="password"
					autocomplete="current-password"
					placeholder="Your recognition phrase…"
					required
					disabled={loading}
				/>
			</div>

			<button type="submit" class="btn" disabled={loading}>
				{loading ? 'Signing in…' : 'Sign in'}
			</button>
		</form>

		<p class="back-link"><a href="/">← Back to vestibule</a></p>
	</div>
</div>

<style>
	:global(*, *::before, *::after) { box-sizing: border-box; margin: 0; padding: 0; }
	:global(html, body) {
		height: 100%;
		font-family: var(--font-body);
		background: var(--color-bg);
		color: var(--color-text);
		-webkit-font-smoothing: antialiased;
	}

	.login-shell {
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-6);
	}

	.login-card {
		width: 100%;
		max-width: 400px;
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-xl);
		padding: var(--space-8);
		box-shadow: var(--shadow-xl);
		display: flex;
		flex-direction: column;
		gap: var(--space-6);
	}

	.login-header {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-4);
		text-align: center;
	}

	.blob {
		width: 56px;
		height: 56px;
		border-radius: 60% 40% 55% 45% / 50% 45% 55% 50%;
		background: linear-gradient(135deg, #4B5FD3, #8B5BD3, #D4895A);
		box-shadow: 0 0 24px rgba(75, 95, 211, 0.25), 0 0 48px rgba(139, 91, 211, 0.12);
		animation: breathe 3s ease-in-out infinite;
	}

	@keyframes breathe {
		0%, 100% { border-radius: 60% 40% 55% 45% / 50% 45% 55% 50%; }
		50%       { border-radius: 50% 50% 45% 55% / 55% 50% 50% 45%; }
	}

	h1 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-2xl);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}

	.subtitle {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		line-height: 1.55;
	}

	.error-banner {
		background: var(--color-error-subtle);
		color: var(--color-error);
		border-radius: var(--radius-md);
		padding: var(--space-3) var(--space-4);
		font-size: var(--text-sm);
	}

	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
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

	input {
		height: 44px;
		padding: 0 var(--space-4);
		background: var(--color-surface-raised);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-md);
		font-family: var(--font-body);
		font-size: var(--text-base);
		color: var(--color-text);
		outline: none;
		width: 100%;
		transition: border-color 100ms, box-shadow 100ms;
	}

	input::placeholder { color: var(--color-text-subtle); }

	input:focus {
		border-color: var(--color-accent);
		box-shadow: 0 0 0 3px var(--color-accent-subtle);
	}

	input:disabled { opacity: 0.5; cursor: not-allowed; }

	.btn {
		height: 44px;
		background: var(--color-accent);
		color: var(--color-accent-text);
		border: none;
		border-radius: var(--radius-md);
		font-family: var(--font-body);
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		cursor: pointer;
		transition: background-color 100ms;
		width: 100%;
	}

	.btn:hover:not(:disabled) { background: var(--color-accent-hover); }
	.btn:disabled { opacity: 0.5; cursor: not-allowed; }

	.back-link {
		text-align: center;
		font-size: var(--text-sm);
	}

	.back-link a {
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.back-link a:hover { color: var(--color-text); }

	@media (prefers-reduced-motion: reduce) {
		.blob { animation: none; }
	}
</style>
