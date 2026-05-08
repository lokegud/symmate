<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	let persona = $derived(data.persona);
</script>

<svelte:head>
	<title>{persona.name}'s Symmate</title>
	<meta name="description" content={persona.description} />
</svelte:head>

<main>
	<div class="vestibule">
		<header>
			<div class="mascot-wrap" aria-hidden="true">
				<div class="mascot-glow"></div>
				<img
					src="/mascot/mascot-idle.svg"
					alt=""
					class="mascot"
					width="160"
					height="190"
				/>
			</div>

			<h1 class="persona-name">{persona.name}</h1>

			{#if persona.host_type}
				<span class="badge">{persona.host_type}</span>
			{/if}
		</header>

		<p class="description">{persona.description}</p>

		{#if persona.contact?.location}
			<p class="location">
				<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
					<path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
				</svg>
				{persona.contact.location}
			</p>
		{/if}

		<a href="/chat" class="cta">Start a conversation →</a>

		{#if persona.capabilities?.length}
			<div class="capabilities">
				{#each persona.capabilities as cap}
					<span class="cap-tag">{cap}</span>
				{/each}
			</div>
		{/if}

		<footer>
			{#if persona.contact?.website}
				<a href={persona.contact.website} target="_blank" rel="noopener noreferrer">
					{persona.contact.website.replace(/^https?:\/\//, '')}
				</a>
			{/if}
			<a href="/.well-known/agent-card.json" target="_blank">agent card</a>
			<a href="/llms.txt">llms.txt</a>
		</footer>
	</div>
</main>

<style>
	:global(body) {
		background: var(--color-bg);
		color: var(--color-text);
		font-family: var(--font-body);
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-6);
	}

	.vestibule {
		max-width: 480px;
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: var(--space-6);
	}

	header {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-3);
		text-align: center;
	}

	/* Mascot */
	.mascot-wrap {
		position: relative;
		width: 160px;
		height: 190px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.mascot-glow {
		position: absolute;
		inset: -20px;
		border-radius: var(--radius-full);
		background: radial-gradient(
			ellipse at 50% 60%,
			color-mix(in srgb, var(--color-accent) 22%, transparent),
			color-mix(in srgb, var(--color-warm) 10%, transparent) 55%,
			transparent 75%
		);
		animation: glow-pulse 3s ease-in-out infinite;
	}

	.mascot {
		position: relative;
		width: 160px;
		height: 190px;
		object-fit: contain;
		animation: mascot-breathe 3s ease-in-out infinite;
		filter: drop-shadow(0 8px 24px color-mix(in srgb, var(--color-accent) 28%, transparent));
	}

	@keyframes mascot-breathe {
		0%, 100% { transform: translateY(0) scale(1); }
		50%       { transform: translateY(-7px) scale(1.025); }
	}

	@keyframes glow-pulse {
		0%, 100% { opacity: 0.7; transform: scale(1); }
		50%       { opacity: 1;   transform: scale(1.08); }
	}

	@media (prefers-reduced-motion: reduce) {
		.mascot, .mascot-glow { animation: none; }
	}

	/* Typography */
	.persona-name {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-2xl);
		font-weight: var(--weight-bold);
		letter-spacing: -0.03em;
		line-height: 1.15;
		color: var(--color-text);
	}

	.badge {
		font-family: var(--font-body);
		font-size: var(--text-xs);
		font-weight: var(--weight-medium);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--color-text-subtle);
		background: var(--color-surface-raised);
		border: 1px solid var(--color-border);
		padding: 0.2em 0.65em;
		border-radius: var(--radius-full);
	}

	.description {
		color: var(--color-text-muted);
		line-height: 1.6;
		text-align: center;
		font-size: var(--text-base);
	}

	.location {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-1);
		color: var(--color-text-subtle);
		font-size: var(--text-sm);
	}

	.location svg {
		color: var(--color-text-subtle);
		flex-shrink: 0;
		opacity: 0.7;
	}

	/* CTA — prism gradient (brand moment) */
	.cta {
		display: block;
		background: linear-gradient(
			135deg,
			var(--color-accent),
			color-mix(in srgb, var(--color-accent) 50%, var(--color-warm)),
			var(--color-warm)
		);
		color: #fff;
		text-decoration: none;
		text-align: center;
		padding: var(--space-4) var(--space-6);
		border-radius: var(--radius-lg);
		font-family: var(--font-body);
		font-weight: var(--weight-semibold);
		font-size: var(--text-base);
		letter-spacing: -0.01em;
		transition: opacity 0.15s ease, transform 0.15s ease;
		box-shadow: 0 4px 20px color-mix(in srgb, var(--color-accent) 30%, transparent);
	}

	.cta:hover {
		opacity: 0.92;
		transform: translateY(-1px);
		box-shadow: 0 6px 28px color-mix(in srgb, var(--color-accent) 38%, transparent);
	}

	.cta:active {
		transform: translateY(0);
		opacity: 0.85;
	}

	/* Capabilities */
	.capabilities {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		justify-content: center;
	}

	.cap-tag {
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		color: var(--color-text-muted);
		padding: 0.22em 0.7em;
		border-radius: var(--radius-full);
		transition: border-color 0.12s, color 0.12s;
	}

	.cap-tag:hover {
		border-color: var(--color-border);
		color: var(--color-text);
	}

	/* Footer */
	footer {
		display: flex;
		justify-content: center;
		gap: var(--space-6);
		padding-top: var(--space-3);
		border-top: 1px solid var(--color-border-subtle);
	}

	footer a {
		color: var(--color-text-subtle);
		font-size: var(--text-sm);
		text-decoration: none;
		padding: var(--space-3) var(--space-2);
		min-height: 44px;
		display: flex;
		align-items: center;
		transition: color 0.12s;
	}

	footer a:hover {
		color: var(--color-text-muted);
	}
</style>
