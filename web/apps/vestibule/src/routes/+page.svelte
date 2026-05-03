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
			<div class="avatar-placeholder" aria-hidden="true">
				{persona.name[0]}
			</div>
			<h1>{persona.name}</h1>
			{#if persona.host_type}
				<span class="badge">{persona.host_type}</span>
			{/if}
		</header>

		<p class="description">{persona.description}</p>

		{#if persona.contact?.location}
			<p class="location">
				<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/></svg>
				{persona.contact.location}
			</p>
		{/if}

		<a href="/.well-known/agent-card.json" class="cta" target="_blank">View agent card →</a>

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
	:global(*, *::before, *::after) {
		box-sizing: border-box;
		margin: 0;
		padding: 0;
	}

	:global(body) {
		font-family: system-ui, -apple-system, sans-serif;
		background: #0f0f11;
		color: #e8e8e8;
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
	}

	.vestibule {
		max-width: 480px;
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	header {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		text-align: center;
	}

	.avatar-placeholder {
		width: 72px;
		height: 72px;
		border-radius: 50%;
		background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: 700;
		color: #fff;
		user-select: none;
	}

	h1 {
		font-size: 1.75rem;
		font-weight: 700;
		letter-spacing: -0.02em;
	}

	.badge {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #9ca3af;
		background: #1e1e24;
		padding: 0.2rem 0.6rem;
		border-radius: 99px;
		border: 1px solid #2e2e38;
	}

	.description {
		color: #b0b0c0;
		line-height: 1.6;
		text-align: center;
		font-size: 0.95rem;
	}

	.location {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.3rem;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.location svg {
		color: #4b5563;
		flex-shrink: 0;
	}

	.cta {
		display: block;
		background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
		color: #fff;
		text-decoration: none;
		text-align: center;
		padding: 0.9rem 1.5rem;
		border-radius: 12px;
		font-weight: 600;
		font-size: 1rem;
		letter-spacing: -0.01em;
		transition: opacity 0.15s;
	}

	.cta:hover {
		opacity: 0.9;
	}

	.capabilities {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
		justify-content: center;
	}

	.cap-tag {
		font-size: 0.75rem;
		background: #1e1e24;
		border: 1px solid #2e2e38;
		padding: 0.2rem 0.65rem;
		border-radius: 99px;
		color: #9ca3af;
	}

	footer {
		display: flex;
		justify-content: center;
		gap: 1.25rem;
		padding-top: 0.5rem;
		border-top: 1px solid #1e1e24;
	}

	footer a {
		color: #6b7280;
		font-size: 0.8rem;
		text-decoration: none;
		padding: 0.75rem 0.5rem;
		min-height: 44px;
		display: flex;
		align-items: center;
	}

	footer a:hover {
		color: #9ca3af;
	}
</style>
