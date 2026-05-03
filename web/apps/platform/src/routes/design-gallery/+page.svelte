<!--
  Symmate Design Gallery — /design-gallery
  Preview all design system tokens + components.
  Dev-only route — not linked in production nav.
  Wren · 2026-05-02
-->
<script lang="ts">
	import { Button, Card, Badge, Input, Avatar, MascotBlob } from '@symmate/design-system';
	import type { BlobState } from '@symmate/design-system';

	let theme = $state<'light' | 'dark'>('light');
	let inputValue = $state('');
	let selectedBlobState = $state<BlobState>('idle');
	let blobEvolved = $state(false);

	const blobStates: BlobState[] = [
		'idle', 'thinking', 'listening', 'happy', 'content',
		'nervous', 'reluctant', 'guarding', 'disagreeing', 'dormant',
	];

	function toggleTheme() {
		theme = theme === 'light' ? 'dark' : 'light';
	}
</script>

<div data-theme={theme} class="gallery">
	<header class="gallery__header">
		<div class="gallery__header-inner">
			<h1 class="gallery__title font-display">Symmate Design Gallery</h1>
			<p class="gallery__subtitle">Design System v1 · Wren · web/design-system branch</p>
			<Button variant="secondary" size="sm" onclick={toggleTheme}>
				{theme === 'light' ? '◐ Dark mode' : '◑ Light mode'}
			</Button>
		</div>
	</header>

	<main class="gallery__main">

		<!-- ── Typography ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Typography</h2>

			<div class="type-specimen">
				<div class="type-row">
					<span class="type-label">Display / Fraunces</span>
					<p class="font-display" style="font-size: var(--text-5xl); line-height: var(--leading-tight); letter-spacing: var(--tracking-tight);">
						Your Symmate<br><em>remembers.</em>
					</p>
				</div>
				<div class="type-row">
					<span class="type-label">Hero — 4xl</span>
					<p class="font-display" style="font-size: var(--text-4xl); line-height: var(--leading-snug);">Like a Tamagotchi that will fix your website.</p>
				</div>
				<div class="type-row">
					<span class="type-label">Body — Plus Jakarta Sans</span>
					<p style="font-size: var(--text-base); line-height: var(--leading-normal); max-width: 60ch;">
						Your Symmate is a private AI representative — one per person, one per business.
						It negotiates on your behalf, guards your context, and has the standing to say no
						when something doesn't feel right. The more you work with it, the more it knows you.
					</p>
				</div>
				<div class="type-row">
					<span class="type-label">Mono — JetBrains Mono</span>
					<p style="font-family: var(--font-mono); font-size: var(--text-sm);">
						GET /agent-card.json · 200 OK · 12ms<br>
						A2A agent-card at loke.symmate.com/.well-known/agent-card.json
					</p>
				</div>
			</div>
		</section>

		<!-- ── Color ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Color</h2>

			<div class="color-grid">
				{#each [
					{ label: 'Background', var: '--color-bg' },
					{ label: 'Surface', var: '--color-surface', border: true },
					{ label: 'Surface Raised', var: '--color-surface-raised', border: true },
					{ label: 'Accent', var: '--color-accent', dark: true },
					{ label: 'Accent Subtle', var: '--color-accent-subtle' },
					{ label: 'Warm', var: '--color-warm', dark: true },
					{ label: 'Warm Subtle', var: '--color-warm-subtle' },
					{ label: 'Success', var: '--color-success', dark: true },
					{ label: 'Warning', var: '--color-warning', dark: true },
					{ label: 'Error', var: '--color-error', dark: true },
					{ label: 'Text', var: '--color-text', dark: true },
					{ label: 'Text Muted', var: '--color-text-muted', dark: true },
					{ label: 'Border', var: '--color-border' },
				] as swatch}
					<div
						class="color-swatch"
						class:color-swatch--bordered={swatch.border}
						style="background: var({swatch.var})"
					>
						<span class="color-swatch__label" class:color-swatch__label--dark={swatch.dark}>
							{swatch.label}
						</span>
					</div>
				{/each}

				<div class="color-swatch color-swatch--prism">
					<span class="color-swatch__label color-swatch__label--dark">Prism gradient</span>
				</div>
			</div>
		</section>

		<!-- ── Mascot Blob ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">MascotBlob</h2>

			<div class="blob-controls">
				<div class="blob-state-picker">
					{#each blobStates as s}
						<button
							class="blob-state-btn"
							class:blob-state-btn--active={selectedBlobState === s}
							onclick={() => selectedBlobState = s}
						>{s}</button>
					{/each}
				</div>
				<label class="evolved-toggle">
					<input type="checkbox" bind:checked={blobEvolved} />
					Evolved (dense marks)
				</label>
			</div>

			<div class="blob-showcase">
				<!-- Platform canonical -->
				<div class="blob-item">
					<MascotBlob seed="platform-canonical" state={selectedBlobState} size="xl" evolved={blobEvolved} />
					<p class="blob-item__label">platform-canonical</p>
					<Badge variant="neutral">{selectedBlobState}</Badge>
				</div>

				<!-- Loke's seeded variant -->
				<div class="blob-item">
					<MascotBlob seed="loke-symmate-v1" state={selectedBlobState} size="xl" evolved={blobEvolved} />
					<p class="blob-item__label">loke-symmate-v1</p>
					<Badge variant="info">{selectedBlobState}</Badge>
				</div>

				<!-- Size scale -->
				<div class="blob-item blob-item--sizes">
					{#each (['sm', 'md', 'lg', 'xl'] as const) as sz}
						<div class="blob-size-row">
							<MascotBlob seed="platform-canonical" state="idle" size={sz} />
							<span class="blob-size-label">{sz}</span>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<!-- ── Buttons ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Buttons</h2>

			<div class="component-row">
				<Button variant="primary">Ask Symmate</Button>
				<Button variant="secondary">Settings</Button>
				<Button variant="ghost">Learn more</Button>
				<Button variant="warm">Get started</Button>
				<Button variant="danger">Delete instance</Button>
			</div>
			<div class="component-row">
				<Button variant="primary" size="sm">Small</Button>
				<Button variant="primary" size="md">Medium</Button>
				<Button variant="primary" size="lg">Large</Button>
				<Button variant="primary" loading>Loading</Button>
				<Button variant="primary" disabled>Disabled</Button>
			</div>
		</section>

		<!-- ── Badges ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Badges</h2>

			<div class="component-row">
				<Badge variant="success" dot>Active</Badge>
				<Badge variant="warning" dot>Degraded</Badge>
				<Badge variant="error" dot>Offline</Badge>
				<Badge variant="info">A2A</Badge>
				<Badge variant="neutral">Draft</Badge>
				<Badge variant="prism">Symmate</Badge>
			</div>
		</section>

		<!-- ── Input ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Input</h2>

			<div style="max-width: 400px; display: flex; flex-direction: column; gap: var(--stack-md);">
				<Input
					label="Display name"
					placeholder="Loke's Symmate"
					hint="This appears on your vestibule."
					bind:value={inputValue}
				/>
				<Input
					label="API endpoint"
					placeholder="https://loke.symmate.com"
					required
				/>
				<Input
					label="Email (error state)"
					placeholder="you@example.com"
					error="Enter a valid email address."
				/>
			</div>
		</section>

		<!-- ── Cards ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Cards</h2>

			<div class="card-row">
				<Card variant="elevated" padding="lg">
					<h3 style="font-size: var(--text-lg); margin: 0 0 var(--stack-sm)">Elevated card</h3>
					<p style="color: var(--color-text-muted); margin: 0; font-size: var(--text-sm);">
						For key content sections. Shadow + border.
					</p>
				</Card>
				<Card variant="outlined" padding="lg">
					<h3 style="font-size: var(--text-lg); margin: 0 0 var(--stack-sm)">Outlined card</h3>
					<p style="color: var(--color-text-muted); margin: 0; font-size: var(--text-sm);">
						For secondary content. Border only.
					</p>
				</Card>
				<Card variant="flat" padding="lg">
					<h3 style="font-size: var(--text-lg); margin: 0 0 var(--stack-sm)">Flat card</h3>
					<p style="color: var(--color-text-muted); margin: 0; font-size: var(--text-sm);">
						For grouped list items. Tinted background.
					</p>
				</Card>
				<Card variant="glass" padding="lg">
					<h3 style="font-size: var(--text-lg); margin: 0 0 var(--stack-sm)">Glass card</h3>
					<p style="color: var(--color-text-muted); margin: 0; font-size: var(--text-sm);">
						Vestibule overlays. Backdrop blur.
					</p>
				</Card>
			</div>
		</section>

		<!-- ── Avatar ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Avatar</h2>

			<div class="component-row" style="align-items: center;">
				<Avatar name="Loke Gudjonsson" size="xs" />
				<Avatar name="Loke Gudjonsson" size="sm" />
				<Avatar name="Loke Gudjonsson" size="md" />
				<Avatar name="Loke Gudjonsson" size="lg" />
				<Avatar name="Loke Gudjonsson" size="xl" />
				<Avatar name="Loke Gudjonsson" size="lg" ring ringColor="prism" />
				<Avatar name="Loke Gudjonsson" size="lg" ring ringColor="accent" />
			</div>
		</section>

		<!-- ── Spacing scale ── -->
		<section class="gallery__section">
			<h2 class="gallery__section-title">Spacing scale</h2>
			<div class="spacing-scale">
				{#each [
					{ label: '1 (4px)',   size: '4px' },
					{ label: '2 (8px)',   size: '8px' },
					{ label: '3 (12px)', size: '12px' },
					{ label: '4 (16px)', size: '16px' },
					{ label: '6 (24px)', size: '24px' },
					{ label: '8 (32px)', size: '32px' },
					{ label: '12 (48px)', size: '48px' },
					{ label: '16 (64px)', size: '64px' },
				] as step}
					<div class="spacing-row">
						<span class="spacing-label">{step.label}</span>
						<div class="spacing-bar" style="width: {step.size}; height: 20px; background: var(--color-accent-subtle); border: 1px solid var(--color-accent);"></div>
					</div>
				{/each}
			</div>
		</section>

	</main>
</div>

<style>
	.gallery {
		min-height: 100vh;
		background: var(--color-bg);
		color: var(--color-text);
		font-family: var(--font-body);
	}

	.gallery__header {
		background: var(--color-surface);
		border-bottom: 1px solid var(--color-border);
		padding: var(--inset-lg) var(--inset-xl);
	}

	.gallery__header-inner {
		max-width: var(--width-dashboard);
		margin: 0 auto;
		display: flex;
		align-items: center;
		gap: var(--space-4);
		flex-wrap: wrap;
	}

	.gallery__title {
		font-size: var(--text-2xl);
		font-weight: var(--weight-bold);
		margin: 0;
		flex: 1;
	}

	.gallery__subtitle {
		color: var(--color-text-muted);
		font-size: var(--text-sm);
		margin: 0;
	}

	.gallery__main {
		max-width: var(--width-dashboard);
		margin: 0 auto;
		padding: var(--space-8) var(--inset-xl);
		display: flex;
		flex-direction: column;
		gap: var(--space-16);
	}

	.gallery__section {
		display: flex;
		flex-direction: column;
		gap: var(--stack-lg);
	}

	.gallery__section-title {
		font-size: var(--text-xl);
		font-weight: var(--weight-semibold);
		margin: 0;
		padding-bottom: var(--stack-sm);
		border-bottom: 1px solid var(--color-border);
	}

	/* Type specimen */
	.type-specimen { display: flex; flex-direction: column; gap: var(--stack-lg); }
	.type-row { display: flex; flex-direction: column; gap: var(--stack-xs); }
	.type-label { font-size: var(--text-xs); letter-spacing: var(--tracking-widest); text-transform: uppercase; color: var(--color-text-subtle); }

	/* Color */
	.color-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
		gap: var(--space-3);
	}
	.color-swatch {
		height: 72px;
		border-radius: var(--radius-md);
		display: flex;
		align-items: flex-end;
		padding: var(--space-2);
	}
	.color-swatch--bordered { border: 1px solid var(--color-border); }
	.color-swatch--prism {
		background: var(--gradient-prism);
	}
	.color-swatch__label {
		font-size: var(--text-xs);
		color: var(--color-text-muted);
		background: var(--color-surface-overlay);
		padding: 2px 4px;
		border-radius: 4px;
	}
	.color-swatch__label--dark { color: rgba(255,255,255,0.9); background: rgba(0,0,0,0.3); }

	/* Blob */
	.blob-controls {
		display: flex;
		flex-direction: column;
		gap: var(--stack-sm);
	}
	.blob-state-picker {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1-5);
	}
	.blob-state-btn {
		padding: var(--space-1) var(--space-2);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-sm);
		background: var(--color-surface);
		color: var(--color-text-muted);
		font-size: var(--text-sm);
		cursor: pointer;
		transition: all var(--duration-micro) var(--ease-out);
	}
	.blob-state-btn--active {
		border-color: var(--color-accent);
		background: var(--color-accent-subtle);
		color: var(--color-accent);
	}
	.evolved-toggle {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		font-size: var(--text-sm);
		cursor: pointer;
	}
	.blob-showcase {
		display: flex;
		align-items: flex-start;
		gap: var(--space-12);
		flex-wrap: wrap;
	}
	.blob-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--stack-sm);
	}
	.blob-item__label {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		font-family: var(--font-mono);
		margin: 0;
	}
	.blob-item--sizes {
		display: flex;
		flex-direction: row;
		align-items: flex-end;
		gap: var(--space-4);
	}
	.blob-size-row {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-2);
	}
	.blob-size-label {
		font-size: var(--text-xs);
		color: var(--color-text-subtle);
	}

	/* Generic rows */
	.component-row {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-3);
		align-items: center;
	}
	.card-row {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: var(--space-4);
	}

	/* Spacing scale */
	.spacing-scale { display: flex; flex-direction: column; gap: var(--space-2); }
	.spacing-row   { display: flex; align-items: center; gap: var(--space-4); }
	.spacing-label { font-size: var(--text-sm); color: var(--color-text-muted); width: 90px; font-family: var(--font-mono); }
</style>
