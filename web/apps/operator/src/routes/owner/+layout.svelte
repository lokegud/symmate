<script lang="ts">
	import '@symmate/design-system/minimal.css';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: any } = $props();

	const navLinks = [
		{ href: '/owner', label: 'Dashboard', icon: 'home' },
		{ href: '/owner/personality', label: 'Personality', icon: 'brain' },
		{ href: '/owner/policy', label: 'Policy', icon: 'shield' },
		{ href: '/owner/users', label: 'Users', icon: 'users' },
		{ href: '/owner/audit', label: 'Audit', icon: 'file-text' },
		{ href: '/owner/conversations', label: 'Conversations', icon: 'messages-square' }
	];
</script>

<svelte:head>
	<title>Owner Dashboard · {data.summary?.name ?? 'Symmate'}</title>
</svelte:head>

<div class="shell">
	<aside class="sidebar">
		<div class="sidebar__brand">
			<a href="/owner" class="brand-link">
				<div class="brand-blob" aria-hidden="true"></div>
				<div>
					<span class="brand-name">{data.summary?.name ?? 'Symmate'}</span>
					<span class="brand-sub">Owner Dashboard</span>
				</div>
			</a>
		</div>

		<nav class="sidebar__nav" aria-label="Owner navigation">
			{#each navLinks as link}
				<a href={link.href} class="nav-item">
					<span class="nav-icon">
						{#if link.icon === 'home'}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
						{:else if link.icon === 'brain'}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M9.5 2a2.5 2.5 0 0 1 5 0v.5a2.5 2.5 0 0 1-5 0V2z"/><path d="M12 2.5V4"/><path d="M7 7.5a5 5 0 0 0 10 0"/><path d="M7 10a5 5 0 0 0 10 0"/><ellipse cx="12" cy="14" rx="7" ry="5"/><path d="M12 19v3"/></svg>
						{:else if link.icon === 'shield'}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
						{:else if link.icon === 'users'}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
						{:else if link.icon === 'file-text'}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
						{:else if link.icon === 'messages-square'}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
						{/if}
					</span>
					{link.label}
				</a>
			{/each}
		</nav>

		<div class="sidebar__footer">
			<a href="/" class="footer-link">← Vestibule</a>
			<form method="POST" action="/owner/logout">
				<button type="submit" class="footer-btn">Sign out</button>
			</form>
		</div>
	</aside>

	<main class="content">
		{@render children()}
	</main>
</div>

<style>
	:global(*, *::before, *::after) { box-sizing: border-box; margin: 0; padding: 0; }
	:global(html) { height: 100%; }
	:global(body) { height: 100%; background: var(--color-bg); color: var(--color-text); }

	.shell {
		display: flex;
		min-height: 100dvh;
	}

	/* Sidebar */
	.sidebar {
		width: 240px;
		flex-shrink: 0;
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		position: sticky;
		top: 0;
		height: 100dvh;
		overflow-y: auto;
	}

	.sidebar__brand {
		padding: var(--space-6);
		border-bottom: 1px solid var(--color-border-subtle);
	}

	.brand-link {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		text-decoration: none;
		color: inherit;
	}

	.brand-blob {
		width: 36px;
		height: 36px;
		border-radius: 60% 40% 55% 45% / 50% 45% 55% 50%;
		background: linear-gradient(135deg, #4B5FD3, #8B5BD3, #D4895A);
		flex-shrink: 0;
		box-shadow: 0 0 16px rgba(75, 95, 211, 0.3);
	}

	.brand-name {
		display: block;
		font-family: var(--font-display);
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		font-optical-sizing: auto;
	}

	.brand-sub {
		display: block;
		font-size: var(--text-xs);
		color: var(--color-text-muted);
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	.sidebar__nav {
		flex: 1;
		padding: var(--space-4) var(--space-3);
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-md);
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		color: var(--color-text-muted);
		text-decoration: none;
		transition: background-color 100ms, color 100ms;
	}

	.nav-item:hover {
		background: var(--color-surface-raised);
		color: var(--color-text);
	}

	:global(.nav-item.active),
	:global(a[aria-current="page"].nav-item) {
		background: var(--color-accent-subtle);
		color: var(--color-accent);
	}

	.nav-icon {
		display: flex;
		align-items: center;
		flex-shrink: 0;
		color: currentColor;
	}

	.sidebar__footer {
		padding: var(--space-4) var(--space-6);
		border-top: 1px solid var(--color-border-subtle);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}

	.footer-link {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.footer-link:hover { color: var(--color-text); }

	.footer-btn {
		background: none;
		border: none;
		font-family: var(--font-body);
		font-size: var(--text-sm);
		color: var(--color-error);
		cursor: pointer;
		padding: 0;
		text-align: left;
	}

	.footer-btn:hover { text-decoration: underline; }

	/* Content area */
	.content {
		flex: 1;
		overflow-y: auto;
		padding: var(--space-8);
		max-width: var(--width-dashboard);
	}

	@media (max-width: 768px) {
		.sidebar { display: none; }
		.content { padding: var(--space-4); }
	}
</style>
