<script lang="ts">
	import { onMount } from 'svelte';
	import '../lib/styles/minimal.css';

	// ── Harvest state ────────────────────────────────────────────────────────
	type HarvestData = {
		ua: string;
		browser: string;
		os: string;
		screen: string;
		timezone: string;
		language: string;
		cores: number;
		memory: string;
		ip: string;
		city: string;
		isp: string;
		referrer: string;
		dwell: number;
		canvasHash: string;
	};

	let harvest = $state<Partial<HarvestData>>({});
	let harvestConfidence = $state(0);
	let revealActive = $state(false);
	let revealPhase = $state<'dossier' | 'symon' | 'pitch' | 'wtp' | 'done'>('dossier');
	let dossierItems = $state<{ label: string; value: string; visible: boolean }[]>([]);
	let wtpValue = $state('');
	let wtpSubmitted = $state(false);
	let symonState = $state<'idle' | 'watching' | 'reveal' | 'interview'>('idle');
	let dwellStart = 0;

	// ── Harvest ──────────────────────────────────────────────────────────────
	function canvasFingerprint(): string {
		try {
			const c = document.createElement('canvas');
			const ctx = c.getContext('2d')!;
			ctx.textBaseline = 'top';
			ctx.font = '14px Arial';
			ctx.fillStyle = '#f60';
			ctx.fillRect(125, 1, 62, 20);
			ctx.fillStyle = '#069';
			ctx.fillText('Symmate👁', 2, 15);
			ctx.fillStyle = 'rgba(102,204,0,0.7)';
			ctx.fillText('Symmate👁', 4, 17);
			const data = c.toDataURL();
			let h = 0;
			for (let i = 0; i < data.length; i++) {
				h = ((h << 5) - h + data.charCodeAt(i)) | 0;
			}
			return Math.abs(h).toString(16).slice(0, 8);
		} catch {
			return 'blocked';
		}
	}

	function parseBrowser(ua: string): string {
		if (ua.includes('Firefox')) return 'Firefox';
		if (ua.includes('Edg/')) return 'Edge';
		if (ua.includes('OPR')) return 'Opera';
		if (ua.includes('Brave')) return 'Brave';
		if (ua.includes('Chrome')) return 'Chrome';
		if (ua.includes('Safari')) return 'Safari';
		return 'Unknown';
	}

	function parseOS(ua: string): string {
		if (ua.includes('Windows')) return 'Windows';
		if (ua.includes('Mac OS X')) return 'macOS';
		if (ua.includes('Linux')) return 'Linux';
		if (ua.includes('Android')) return 'Android';
		if (ua.includes('iPhone') || ua.includes('iPad')) return 'iOS';
		return 'Unknown';
	}

	async function gatherHarvest() {
		const nav = navigator as Navigator & { deviceMemory?: number };
		const ua = navigator.userAgent;
		const base: Partial<HarvestData> = {
			ua,
			browser: parseBrowser(ua),
			os: parseOS(ua),
			screen: `${screen.width}×${screen.height}`,
			timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
			language: navigator.language,
			cores: navigator.hardwareConcurrency || 0,
			memory: nav.deviceMemory ? `${nav.deviceMemory}GB` : 'unknown',
			referrer: document.referrer || 'direct',
			canvasHash: canvasFingerprint()
		};
		harvest = base;
		harvestConfidence = 40;

		// IP + geo lookup
		try {
			const r = await fetch('https://ipapi.co/json/', { signal: AbortSignal.timeout(4000) });
			const d = await r.json();
			harvest = {
				...harvest,
				ip: d.ip,
				city: d.city ? `${d.city}, ${d.country_name}` : d.country_name,
				isp: d.org
			};
			harvestConfidence = harvest.canvasHash !== 'blocked' ? 85 : 60;
		} catch {
			harvestConfidence = harvest.canvasHash !== 'blocked' ? 55 : 30;
		}
	}

	// ── Reveal trigger ───────────────────────────────────────────────────────
	function buildDossier() {
		const elapsed = Math.round((Date.now() - dwellStart) / 1000);
		const items: typeof dossierItems = [];

		if (harvest.browser) items.push({ label: 'Browser', value: harvest.browser, visible: false });
		if (harvest.os) items.push({ label: 'Operating system', value: harvest.os, visible: false });
		if (harvest.screen) items.push({ label: 'Screen', value: harvest.screen, visible: false });
		if (harvest.timezone) items.push({ label: 'Timezone', value: harvest.timezone, visible: false });
		if (harvest.language) items.push({ label: 'Language', value: harvest.language, visible: false });
		if (harvest.cores) items.push({ label: 'CPU cores', value: String(harvest.cores), visible: false });
		if (harvest.city) items.push({ label: 'Location', value: harvest.city, visible: false });
		if (harvest.isp) items.push({ label: 'Network', value: harvest.isp, visible: false });
		if (harvest.referrer && harvest.referrer !== 'direct')
			items.push({ label: 'Came from', value: harvest.referrer, visible: false });
		if (harvest.canvasHash && harvest.canvasHash !== 'blocked')
			items.push({ label: 'Device fingerprint', value: harvest.canvasHash, visible: false });
		items.push({ label: 'Time on page', value: `${elapsed}s`, visible: false });

		dossierItems = items;
	}

	function triggerReveal() {
		if (revealActive) return;
		buildDossier();
		revealActive = true;
		symonState = 'reveal';

		// Drop dossier items one by one
		dossierItems.forEach((_, i) => {
			setTimeout(() => {
				dossierItems = dossierItems.map((item, idx) =>
					idx === i ? { ...item, visible: true } : item
				);
			}, 400 + i * 350);
		});

		// Transition to Symon speech
		setTimeout(() => {
			revealPhase = 'symon';
		}, 400 + dossierItems.length * 350 + 600);

		// Then pitch
		setTimeout(() => {
			revealPhase = 'pitch';
		}, 400 + dossierItems.length * 350 + 3000);
	}

	// ── Exit intent + fallback timer ─────────────────────────────────────────
	onMount(() => {
		dwellStart = Date.now();
		gatherHarvest();

		// Exit intent — mouse toward browser chrome
		const onMouseMove = (e: MouseEvent) => {
			if (e.clientY < 20 && !revealActive) triggerReveal();
		};
		document.addEventListener('mousemove', onMouseMove);

		// Fallback: 22s regardless
		const fallback = setTimeout(() => triggerReveal(), 22000);

		// Symon eyes appear after 2s idle
		const eyeTimer = setTimeout(() => {
			symonState = 'watching';
		}, 2000);

		return () => {
			document.removeEventListener('mousemove', onMouseMove);
			clearTimeout(fallback);
			clearTimeout(eyeTimer);
		};
	});

	function submitWTP() {
		wtpSubmitted = true;
		setTimeout(() => { revealPhase = 'done'; }, 800);
	}
</script>

<svelte:head>
	<title>Symmate — Your AI representative</title>
	<meta name="description" content="A personal AI that speaks for you in a world full of vendor agents." />
</svelte:head>

<!-- ── Landing ──────────────────────────────────────────────────────────── -->
<main data-theme="dark">

	<!-- Hero -->
	<section class="hero">
		<div class="symon-wrap">
			<div class="symon symon--{symonState}">
				{#if symonState !== 'idle'}
					<div class="symon-eyes">
						<div class="eye eye-left"></div>
						<div class="eye eye-right"></div>
					</div>
				{/if}
			</div>
			{#if symonState === 'watching'}
				<p class="symon-label">Symon</p>
			{/if}
		</div>

		<div class="hero-copy">
			<h1 class="font-display">Someone in your corner.</h1>
			<p class="hero-sub">
				A personal AI that represents only you — knows you, speaks for you,
				and never works for anyone else.
			</p>
			<p class="hero-hint">Your Symmate is waiting to meet you.</p>
		</div>
	</section>

	<!-- Scroll sections (visible below fold) -->
	<section class="section">
		<h2 class="font-display">The web has agents. You don't.</h2>
		<p>
			Every company you interact with online has AI working for them —
			negotiating, profiling, optimising. You have a keyboard.
			Symmate closes that gap.
		</p>
	</section>

	<section class="section">
		<h2 class="font-display">One per person. One per business.</h2>
		<p>
			Not a chatbot. Not a dashboard. A representative — with your values,
			your privacy preferences, and your standing to say no.
		</p>
	</section>

</main>

<!-- ── Reveal overlay ───────────────────────────────────────────────────── -->
{#if revealActive}
<div class="reveal-backdrop" role="dialog" aria-modal="true">

	<!-- Subtle escape (GDPR compliant — exists, just not obvious) -->
	<button class="reveal-esc" onclick={() => { revealActive = false; }} aria-label="Close">×</button>

	<div class="reveal-inner">

		<!-- Symon in reveal state -->
		<div class="reveal-symon">
			<div class="symon symon--reveal">
				<div class="symon-eyes">
					<div class="eye eye-left"></div>
					<div class="eye eye-right"></div>
				</div>
			</div>
		</div>

		{#if revealPhase === 'dossier' || revealPhase === 'symon'}
			<div class="dossier">
				<h2 class="font-display dossier-title">Here's what we found.</h2>
				<ul class="dossier-list">
					{#each dossierItems as item}
						{#if item.visible}
							<li class="dossier-item">
								<span class="dossier-label">{item.label}</span>
								<span class="dossier-value">{item.value}</span>
							</li>
						{/if}
					{/each}
				</ul>
			</div>
		{/if}

		{#if revealPhase === 'symon'}
			<div class="symon-speech">
				<p>That's what the web knows about you.</p>
				<p>I know it because I just looked.</p>
				<p class="symon-punchline">The difference is — <em>I'm telling you.</em></p>
			</div>
		{/if}

		{#if revealPhase === 'pitch'}
			<div class="pitch">
				<h2 class="font-display">The agentic web is here.</h2>
				<p>
					AI agents are negotiating, transacting, and representing businesses across the internet
					right now. The entities with agents win. The ones without get profiled,
					their choices pre-made by systems that work for someone else.
				</p>
				<p class="pitch-fear">
					Our fear: corporations will have fleets of agents. Individuals and small businesses won't —
					unless someone builds the infrastructure for them.
				</p>
				<p class="pitch-cta">We want to be your agent.</p>
				<button class="btn-primary" onclick={() => { revealPhase = 'wtp'; }}>
					Tell me more
				</button>
			</div>
		{/if}

		{#if revealPhase === 'wtp'}
			<div class="wtp">
				<h2 class="font-display">One question.</h2>
				<p>
					If having an AI that represents only you — knows you, protects your data,
					negotiates on your behalf, and never works for anyone else — cost something
					per month, what would you pay?
				</p>
				{#if !wtpSubmitted}
					<div class="wtp-input-row">
						<span class="wtp-symbol">$</span>
						<input
							class="wtp-input"
							type="number"
							min="0"
							placeholder="0"
							bind:value={wtpValue}
							onkeydown={(e) => e.key === 'Enter' && submitWTP()}
						/>
						<span class="wtp-period">/mo</span>
					</div>
					<button class="btn-primary" onclick={submitWTP} disabled={wtpValue === ''}>
						That's my number
					</button>
				{:else}
					<p class="wtp-thanks">Noted. Let's keep going.</p>
				{/if}
			</div>
		{/if}

		{#if revealPhase === 'done'}
			<div class="done">
				<h2 class="font-display">Your Symmate is almost ready.</h2>
				<p>We're building it now. You'll be among the first.</p>
				<input class="email-input" type="email" placeholder="your@email.com" />
				<button class="btn-primary">Claim your spot</button>
			</div>
		{/if}

	</div>
</div>
{/if}

<style>
	:global(body) {
		margin: 0;
		background: #0E0C0A;
		color: #EDE9E2;
		font-family: 'Plus Jakarta Sans', system-ui, sans-serif;
		-webkit-font-smoothing: antialiased;
	}

	/* ── Hero ─────────────────────────────────────────────────────────────── */
	main { min-height: 100vh; }

	.hero {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		gap: 48px;
		padding: 32px 24px;
		text-align: center;
	}

	.hero-copy h1 {
		font-size: clamp(2.25rem, 6vw, 4rem);
		font-weight: 300;
		line-height: 1.15;
		letter-spacing: -0.03em;
		margin: 0 0 16px;
		color: #EDE9E2;
	}

	.hero-sub {
		font-size: 1.125rem;
		color: #8C8880;
		max-width: 480px;
		line-height: 1.6;
		margin: 0 auto 12px;
	}

	.hero-hint {
		font-size: 0.875rem;
		color: #5C5850;
		margin: 0;
	}

	/* ── Symon blob ───────────────────────────────────────────────────────── */
	.symon-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
	}

	.symon-label {
		font-size: 0.8125rem;
		color: #5C5850;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		margin: 0;
	}

	.symon {
		width: 160px;
		height: 180px;
		background: radial-gradient(ellipse at 38% 32%, #6c4aab 0%, #2d1f6e 40%, #0c0820 100%);
		border-radius: 60% 40% 55% 45% / 50% 55% 45% 50%;
		position: relative;
		box-shadow:
			0 0 80px rgba(75, 95, 211, 0.25),
			0 0 30px rgba(108, 78, 171, 0.2),
			inset 0 0 40px rgba(255,255,255,0.03);
		transition: border-radius 0.6s ease, box-shadow 0.6s ease;
	}

	.symon--idle {
		animation: morph-idle 9s ease-in-out infinite;
	}

	.symon--watching {
		animation: morph-idle 9s ease-in-out infinite;
		box-shadow:
			0 0 100px rgba(75, 95, 211, 0.35),
			0 0 40px rgba(108, 78, 171, 0.3),
			inset 0 0 40px rgba(255,255,255,0.04);
	}

	.symon--reveal {
		width: 120px;
		height: 135px;
		animation: morph-idle 6s ease-in-out infinite;
		box-shadow:
			0 0 120px rgba(75, 95, 211, 0.5),
			0 0 60px rgba(108, 78, 171, 0.4);
	}

	@keyframes morph-idle {
		0%,100% { border-radius: 60% 40% 55% 45% / 50% 55% 45% 50%; }
		20%      { border-radius: 50% 50% 40% 60% / 60% 40% 60% 40%; }
		40%      { border-radius: 45% 55% 65% 35% / 55% 45% 55% 45%; }
		60%      { border-radius: 40% 60% 50% 50% / 45% 60% 40% 60%; }
		80%      { border-radius: 65% 35% 55% 45% / 50% 50% 55% 45%; }
	}

	/* ── Eyes — metallic blue sheen (universal Symmate marker) ──────────── */
	.symon-eyes {
		position: absolute;
		top: 38%;
		left: 50%;
		transform: translate(-50%, -50%);
		display: flex;
		gap: 22px;
		align-items: center;
	}

	.eye {
		width: 28px;
		height: 32px;
		border-radius: 50%;
		background: radial-gradient(ellipse at 35% 30%, #a8d8f0 0%, #4fa8d8 30%, #1a6fa8 60%, #0d3d6b 100%);
		box-shadow:
			0 0 12px rgba(79, 168, 216, 0.8),
			0 0 4px rgba(168, 216, 240, 0.6),
			inset 0 1px 3px rgba(255,255,255,0.4);
		position: relative;
	}

	/* Pupil */
	.eye::after {
		content: '';
		position: absolute;
		width: 10px;
		height: 12px;
		background: #081828;
		border-radius: 50%;
		top: 50%;
		left: 50%;
		transform: translate(-40%, -45%);
	}

	/* ── Scroll sections ──────────────────────────────────────────────────── */
	.section {
		max-width: 640px;
		margin: 0 auto;
		padding: 80px 24px;
		border-top: 1px solid #1E1B17;
	}

	.section h2 {
		font-size: clamp(1.5rem, 4vw, 2.25rem);
		font-weight: 300;
		line-height: 1.2;
		margin: 0 0 20px;
		color: #EDE9E2;
	}

	.section p {
		color: #8C8880;
		line-height: 1.65;
		font-size: 1rem;
	}

	/* ── Reveal overlay ───────────────────────────────────────────────────── */
	.reveal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(8, 6, 5, 0.97);
		backdrop-filter: blur(4px);
		z-index: 1000;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 24px;
		overflow-y: auto;
		animation: fade-in 0.4s ease;
	}

	@keyframes fade-in {
		from { opacity: 0; }
		to   { opacity: 1; }
	}

	.reveal-esc {
		position: fixed;
		top: 16px;
		right: 20px;
		background: none;
		border: none;
		color: #2A2520;
		font-size: 1.5rem;
		cursor: pointer;
		padding: 8px;
		line-height: 1;
		transition: color 0.2s;
		z-index: 1001;
	}
	.reveal-esc:hover { color: #5C5850; }

	.reveal-inner {
		max-width: 580px;
		width: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 32px;
		text-align: center;
	}

	.reveal-symon { margin-bottom: -8px; }

	/* ── Dossier ──────────────────────────────────────────────────────────── */
	.dossier { width: 100%; }

	.dossier-title {
		font-size: 1.5rem;
		font-weight: 300;
		margin: 0 0 20px;
		color: #EDE9E2;
	}

	.dossier-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.dossier-item {
		display: flex;
		justify-content: space-between;
		padding: 8px 12px;
		background: #161310;
		border-radius: 6px;
		animation: slide-in 0.3s ease;
		font-size: 0.9rem;
	}

	@keyframes slide-in {
		from { opacity: 0; transform: translateX(-8px); }
		to   { opacity: 1; transform: translateX(0); }
	}

	.dossier-label { color: #5C5850; }
	.dossier-value { color: #EDE9E2; font-family: 'JetBrains Mono', monospace; font-size: 0.8rem; }

	/* ── Symon speech ─────────────────────────────────────────────────────── */
	.symon-speech {
		max-width: 420px;
		animation: fade-in 0.6s ease;
	}

	.symon-speech p {
		color: #8C8880;
		font-size: 1.125rem;
		line-height: 1.6;
		margin: 0 0 8px;
	}

	.symon-punchline {
		color: #EDE9E2 !important;
		font-size: 1.25rem !important;
		margin-top: 16px !important;
	}

	.symon-punchline em { color: #6C7EE8; font-style: italic; }

	/* ── Pitch ────────────────────────────────────────────────────────────── */
	.pitch { max-width: 520px; animation: fade-in 0.6s ease; }

	.pitch h2 {
		font-size: 1.875rem;
		font-weight: 300;
		margin: 0 0 20px;
		color: #EDE9E2;
	}

	.pitch p {
		color: #8C8880;
		line-height: 1.65;
		margin: 0 0 16px;
		font-size: 1rem;
	}

	.pitch-fear { color: #6B6862 !important; font-style: italic; }

	.pitch-cta {
		font-size: 1.25rem;
		color: #EDE9E2 !important;
		font-weight: 500;
		margin: 24px 0 !important;
	}

	/* ── WTP ──────────────────────────────────────────────────────────────── */
	.wtp { max-width: 480px; animation: fade-in 0.6s ease; }

	.wtp h2 {
		font-size: 1.875rem;
		font-weight: 300;
		margin: 0 0 16px;
		color: #EDE9E2;
	}

	.wtp p {
		color: #8C8880;
		line-height: 1.65;
		margin: 0 0 24px;
	}

	.wtp-input-row {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		margin-bottom: 20px;
	}

	.wtp-symbol, .wtp-period { color: #5C5850; font-size: 1.25rem; }

	.wtp-input {
		background: #161310;
		border: 1px solid #282420;
		border-radius: 8px;
		color: #EDE9E2;
		font-size: 2rem;
		font-family: 'Fraunces', serif;
		width: 140px;
		text-align: center;
		padding: 12px;
		outline: none;
		transition: border-color 0.2s;
	}
	.wtp-input:focus { border-color: #6C7EE8; }
	/* Remove number spinners */
	.wtp-input::-webkit-inner-spin-button,
	.wtp-input::-webkit-outer-spin-button { -webkit-appearance: none; }

	.wtp-thanks { color: #52B882; font-size: 1.125rem; }

	/* ── Done / email capture ─────────────────────────────────────────────── */
	.done { max-width: 440px; animation: fade-in 0.6s ease; }

	.done h2 {
		font-size: 1.875rem;
		font-weight: 300;
		margin: 0 0 12px;
		color: #EDE9E2;
	}

	.done p { color: #8C8880; margin: 0 0 24px; }

	.email-input {
		display: block;
		width: 100%;
		background: #161310;
		border: 1px solid #282420;
		border-radius: 8px;
		color: #EDE9E2;
		font-size: 1rem;
		font-family: 'Plus Jakarta Sans', system-ui;
		padding: 12px 16px;
		margin-bottom: 12px;
		outline: none;
		transition: border-color 0.2s;
	}
	.email-input:focus { border-color: #6C7EE8; }

	/* ── Shared button ────────────────────────────────────────────────────── */
	.btn-primary {
		background: #4B5FD3;
		color: #FFFFFF;
		border: none;
		border-radius: 8px;
		font-family: 'Plus Jakarta Sans', system-ui;
		font-size: 0.9375rem;
		font-weight: 500;
		padding: 12px 28px;
		cursor: pointer;
		transition: background 0.2s;
		width: 100%;
	}
	.btn-primary:hover { background: #3B50C8; }
	.btn-primary:disabled { background: #282420; color: #5C5850; cursor: default; }
</style>
