<script lang="ts">
	import { onMount } from 'svelte';
	import '@symmate/design-system/minimal.css';

	type HarvestData = {
		ua: string; browser: string; os: string; screen: string;
		timezone: string; language: string; cores: number;
		memory: string; ip: string; city: string; isp: string;
		referrer: string; canvasHash: string;
	};

	let harvest = $state<Partial<HarvestData>>({});
	let revealActive   = $state(false);
	let revealPhase    = $state<'dossier' | 'symon' | 'pitch' | 'wtp' | 'done'>('dossier');
	let dossierItems   = $state<{ label: string; value: string; visible: boolean }[]>([]);
	let wtpValue       = $state('');
	let wtpSubmitted   = $state(false);
	let symonState     = $state<'idle' | 'watching' | 'reveal'>('idle');
	let dwellStart     = 0;

	function canvasFingerprint(): string {
		try {
			const c = document.createElement('canvas');
			const ctx = c.getContext('2d')!;
			ctx.font = '14px Arial';
			ctx.fillStyle = '#f60';
			ctx.fillRect(125, 1, 62, 20);
			ctx.fillStyle = '#069';
			ctx.fillText('Symmate', 2, 15);
			let h = 0;
			const data = c.toDataURL();
			for (let i = 0; i < data.length; i++) h = ((h << 5) - h + data.charCodeAt(i)) | 0;
			return Math.abs(h).toString(16).slice(0, 8);
		} catch { return 'blocked'; }
	}

	function parseBrowser(ua: string) {
		if (ua.includes('Firefox')) return 'Firefox';
		if (ua.includes('Edg/'))   return 'Edge';
		if (ua.includes('OPR'))    return 'Opera';
		if (ua.includes('Chrome')) return 'Chrome';
		if (ua.includes('Safari')) return 'Safari';
		return 'Unknown';
	}

	function parseOS(ua: string) {
		if (ua.includes('Windows'))            return 'Windows';
		if (ua.includes('Mac OS X'))           return 'macOS';
		if (ua.includes('Linux'))              return 'Linux';
		if (ua.includes('Android'))            return 'Android';
		if (ua.includes('iPhone') || ua.includes('iPad')) return 'iOS';
		return 'Unknown';
	}

	async function gatherHarvest() {
		const nav = navigator as Navigator & { deviceMemory?: number };
		const ua  = navigator.userAgent;
		harvest = {
			ua, browser: parseBrowser(ua), os: parseOS(ua),
			screen: `${screen.width}×${screen.height}`,
			timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
			language: navigator.language,
			cores: navigator.hardwareConcurrency || 0,
			memory: nav.deviceMemory ? `${nav.deviceMemory}GB` : 'unknown',
			referrer: document.referrer || 'direct',
			canvasHash: canvasFingerprint()
		};
		try {
			const r = await fetch('https://ipapi.co/json/', { signal: AbortSignal.timeout(4000) });
			const d = await r.json();
			harvest = { ...harvest, ip: d.ip, city: d.city ? `${d.city}, ${d.country_name}` : d.country_name, isp: d.org };
		} catch { /* offline ok */ }
	}

	function buildDossier() {
		const elapsed = Math.round((Date.now() - dwellStart) / 1000);
		const items: typeof dossierItems = [];
		if (harvest.browser)  items.push({ label: 'Browser', value: harvest.browser, visible: false });
		if (harvest.os)       items.push({ label: 'OS', value: harvest.os, visible: false });
		if (harvest.screen)   items.push({ label: 'Screen', value: harvest.screen, visible: false });
		if (harvest.timezone) items.push({ label: 'Timezone', value: harvest.timezone, visible: false });
		if (harvest.language) items.push({ label: 'Language', value: harvest.language, visible: false });
		if (harvest.cores)    items.push({ label: 'CPU cores', value: String(harvest.cores), visible: false });
		if (harvest.city)     items.push({ label: 'Location', value: harvest.city, visible: false });
		if (harvest.isp)      items.push({ label: 'Network', value: harvest.isp, visible: false });
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
		dossierItems.forEach((_, i) => {
			setTimeout(() => {
				dossierItems = dossierItems.map((item, idx) => idx === i ? { ...item, visible: true } : item);
			}, 400 + i * 350);
		});
		setTimeout(() => { revealPhase = 'symon'; },  400 + dossierItems.length * 350 + 600);
		setTimeout(() => { revealPhase = 'pitch'; },  400 + dossierItems.length * 350 + 3200);
	}

	onMount(() => {
		dwellStart = Date.now();
		gatherHarvest();
		const onMove = (e: MouseEvent) => { if (e.clientY < 20 && !revealActive) triggerReveal(); };
		document.addEventListener('mousemove', onMove);
		const fallback = setTimeout(() => triggerReveal(), 22000);
		const eyeTimer = setTimeout(() => { symonState = 'watching'; }, 2000);
		return () => {
			document.removeEventListener('mousemove', onMove);
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
	<title>Symmate — Your AI in the agentic web</title>
	<meta name="description" content="A personal AI that represents only you in a world run by agent fleets." />
</svelte:head>

<main>
	<!-- Decorative blobs — crawling around the page -->
	<div class="blob-deco blob-deco-1" aria-hidden="true">
		<img src="/mascot/symon-idle.svg" alt="" width="80" height="88"/>
	</div>
	<div class="blob-deco blob-deco-2" aria-hidden="true">
		<img src="/mascot/symon-listening.svg" alt="" width="56" height="62"/>
	</div>
	<div class="blob-deco blob-deco-3" aria-hidden="true">
		<img src="/mascot/symon-talking.svg" alt="" width="68" height="75"/>
	</div>
	<div class="blob-deco blob-deco-4" aria-hidden="true">
		<img src="/mascot/symon-idle.svg" alt="" width="44" height="48"/>
	</div>
	<div class="blob-deco blob-deco-5" aria-hidden="true">
		<img src="/mascot/symon-listening.svg" alt="" width="52" height="57"/>
	</div>

	<!-- Hero -->
	<section class="hero">
		<div class="symon-wrap">
			<img
				src="/mascot/symon-{symonState === 'idle' ? 'idle' : symonState === 'watching' ? 'listening' : 'talking'}.svg"
				alt="Symon"
				class="symon-hero symon-hero--{symonState}"
				width="180"
				height="198"
			/>
			{#if symonState !== 'idle'}
				<p class="symon-name">Symon</p>
			{/if}
		</div>

		<div class="hero-copy">
			<h1>Someone in your corner.</h1>
			<p class="hero-sub">
				The web has AI agents working for companies. Now you have one working for you.
				Knows you, speaks for you, never works for anyone else.
			</p>
			<a href="#learn-more" class="cta-btn">Meet your Symmate →</a>
		</div>
	</section>

	<!-- Section 1 -->
	<section class="section" id="learn-more">
		<div class="section-blob" aria-hidden="true">
			<img src="/mascot/symon-idle.svg" alt="" width="64" height="70"/>
		</div>
		<h2>The web has agents. You don't. Yet.</h2>
		<p>
			150+ companies are already deploying AI agents that negotiate, profile, and optimise — for themselves.
			Every time you click "accept," something smarter than a cookie is watching.
			Symmate puts an agent in your corner.
		</p>
	</section>

	<!-- Section 2 -->
	<section class="section section-alt">
		<div class="section-blob section-blob-right" aria-hidden="true">
			<img src="/mascot/symon-talking.svg" alt="" width="64" height="70"/>
		</div>
		<h2>One per person. One per business.</h2>
		<p>
			Not a chatbot. Not a subscription to someone else's model. A representative — with your values,
			your privacy preferences, and the standing to say no on your behalf.
			It grows with you. The more you use it, the more it knows you.
		</p>
	</section>

	<!-- Section 3 — push-back feature highlight -->
	<section class="section section-pushback">
		<div class="pushback-pair" aria-hidden="true">
			<img src="/mascot/symon-pushback.svg" alt="" width="80" height="88"/>
		</div>
		<h2>It can say no. That's the point.</h2>
		<p>
			Every AI you've ever used was trained to help you do what you want.
			Your Symmate is the first one trained to stop you when you're about to do something
			you'll regret. Pikachu shocks Ash sometimes. That's what makes it real.
		</p>
	</section>

	<!-- Footer CTA -->
	<section class="footer-cta">
		<img src="/mascot/symon-idle.svg" alt="Symon" width="72" height="79" class="footer-blob"/>
		<h2>Your Symmate is waiting.</h2>
		<p>Join the first wave. Free while we build.</p>
		<div class="email-row">
			<input class="email-input" type="email" placeholder="your@email.com" aria-label="Email address"/>
			<button class="cta-btn" type="button">Claim your spot</button>
		</div>
	</section>
</main>

{#if revealActive}
<div class="reveal-backdrop" role="dialog" aria-modal="true">
	<button class="reveal-esc" onclick={() => { revealActive = false; }} aria-label="Close">×</button>
	<div class="reveal-inner">
		<img
			src="/mascot/symon-{revealPhase === 'dossier' || revealPhase === 'symon' ? 'listening' : revealPhase === 'pitch' ? 'talking' : 'idle'}.svg"
			alt="Symon"
			class="reveal-symon-img"
			width="100"
			height="110"
		/>

		{#if revealPhase === 'dossier' || revealPhase === 'symon'}
			<div class="dossier">
				<h2 class="dossier-title">Here's what we found.</h2>
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
				<h2>The agentic web is here.</h2>
				<p>
					AI agents are negotiating, profiling, and transacting on behalf of companies across the internet
					right now. The entities with agents win. You need one that works for you.
				</p>
				<p class="pitch-cta">We want to be your agent.</p>
				<button class="cta-btn" onclick={() => { revealPhase = 'wtp'; }}>Tell me more</button>
			</div>
		{/if}

		{#if revealPhase === 'wtp'}
			<div class="wtp">
				<h2>One question.</h2>
				<p>
					If your Symmate knew you, protected your data, negotiated on your behalf,
					and never worked for anyone else — what would you pay per month?
				</p>
				{#if !wtpSubmitted}
					<div class="wtp-input-row">
						<span class="wtp-symbol">$</span>
						<input class="wtp-input" type="number" min="0" placeholder="0"
							bind:value={wtpValue} onkeydown={(e) => e.key === 'Enter' && submitWTP()}/>
						<span class="wtp-period">/mo</span>
					</div>
					<button class="cta-btn" onclick={submitWTP} disabled={wtpValue === ''}>That's my number</button>
				{:else}
					<p class="wtp-thanks">Noted. Let's keep going.</p>
				{/if}
			</div>
		{/if}

		{#if revealPhase === 'done'}
			<div class="done">
				<img src="/mascot/symon-idle.svg" alt="Symon" width="80" height="88" class="done-blob"/>
				<h2>Your Symmate is almost ready.</h2>
				<p>We're building it now. You'll be among the first.</p>
				<input class="email-input" type="email" placeholder="your@email.com"/>
				<button class="cta-btn">Claim your spot</button>
			</div>
		{/if}
	</div>
</div>
{/if}

<style>
	:global(html, body) {
		margin: 0;
		background: var(--color-symon-subtle);
		color: var(--color-text);
		font-family: var(--font-body);
		-webkit-font-smoothing: antialiased;
	}

	main {
		min-height: 100vh;
		position: relative;
		overflow: hidden;
	}

	/* ── Decorative crawling blobs ─────────────────────────────────────────── */
	.blob-deco {
		position: fixed;
		pointer-events: none;
		z-index: 0;
		opacity: 0.55;
	}

	.blob-deco-1 { top: 6%;  left: 2%;  animation: blob-drift-1 18s ease-in-out infinite; }
	.blob-deco-2 { top: 22%; right: 4%; animation: blob-drift-2 22s ease-in-out infinite; }
	.blob-deco-3 { top: 58%; left: 5%; animation: blob-drift-3 26s ease-in-out infinite; }
	.blob-deco-4 { top: 78%; right: 2%; animation: blob-drift-4 20s ease-in-out infinite; }
	.blob-deco-5 { top: 45%; right: 8%; animation: blob-drift-5 24s ease-in-out infinite; }

	@keyframes blob-drift-1 {
		0%,100% { transform: translate(0, 0) rotate(0deg); }
		33%      { transform: translate(12px, 18px) rotate(8deg); }
		66%      { transform: translate(-8px, 10px) rotate(-5deg); }
	}
	@keyframes blob-drift-2 {
		0%,100% { transform: translate(0, 0) rotate(0deg); }
		40%      { transform: translate(-14px, 22px) rotate(-10deg); }
		70%      { transform: translate(6px, -8px) rotate(6deg); }
	}
	@keyframes blob-drift-3 {
		0%,100% { transform: translate(0, 0) rotate(0deg); }
		35%      { transform: translate(16px, -12px) rotate(12deg); }
		65%      { transform: translate(-10px, 16px) rotate(-8deg); }
	}
	@keyframes blob-drift-4 {
		0%,100% { transform: translate(0, 0) rotate(0deg); }
		30%      { transform: translate(-18px, -14px) rotate(-14deg); }
		70%      { transform: translate(10px, 8px) rotate(9deg); }
	}
	@keyframes blob-drift-5 {
		0%,100% { transform: translate(0, 0) rotate(0deg); }
		45%      { transform: translate(12px, 20px) rotate(10deg); }
		75%      { transform: translate(-8px, -14px) rotate(-7deg); }
	}

	@media (prefers-reduced-motion: reduce) {
		.blob-deco { animation: none; }
	}

	/* ── Hero ─────────────────────────────────────────────────────────────── */
	.hero {
		position: relative;
		z-index: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		gap: var(--space-12);
		padding: var(--space-8) var(--space-6);
		text-align: center;
	}

	.symon-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-3);
	}

	.symon-hero {
		filter: drop-shadow(0 12px 40px color-mix(in srgb, var(--color-symon) 45%, transparent));
		transition: filter 0.6s ease;
	}

	.symon-hero--watching, .symon-hero--reveal {
		filter: drop-shadow(0 16px 56px color-mix(in srgb, var(--color-symon) 60%, transparent));
		animation: hero-bounce 3s ease-in-out infinite;
	}

	.symon-hero--idle {
		animation: hero-breathe 4s ease-in-out infinite;
	}

	@keyframes hero-breathe {
		0%, 100% { transform: translateY(0) scale(1); }
		50%       { transform: translateY(-8px) scale(1.02); }
	}

	@keyframes hero-bounce {
		0%, 100% { transform: translateY(0) scale(1); }
		50%       { transform: translateY(-10px) scale(1.03); }
	}

	@media (prefers-reduced-motion: reduce) {
		.symon-hero { animation: none !important; }
	}

	.symon-name {
		font-family: var(--font-display);
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		color: var(--color-symon-deep);
		letter-spacing: -0.01em;
		margin: 0;
	}

	.hero-copy h1 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: clamp(var(--text-3xl), 7vw, var(--text-5xl));
		font-weight: var(--weight-bold);
		line-height: 1.1;
		letter-spacing: -0.04em;
		margin: 0 0 var(--space-4);
		color: var(--color-text);
	}

	.hero-sub {
		font-size: clamp(var(--text-base), 2.5vw, var(--text-lg));
		color: var(--color-text-muted);
		max-width: 520px;
		line-height: 1.65;
		margin: 0 auto var(--space-8);
	}

	/* ── CTA button ────────────────────────────────────────────────────────── */
	.cta-btn {
		display: inline-block;
		background: linear-gradient(135deg, var(--color-symon), var(--color-symon-deep));
		color: #fff;
		text-decoration: none;
		border: none;
		border-radius: var(--radius-full);
		font-family: var(--font-body);
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		padding: var(--space-4) var(--space-8);
		cursor: pointer;
		box-shadow: 0 6px 28px color-mix(in srgb, var(--color-symon) 40%, transparent);
		transition: transform 0.15s ease, box-shadow 0.15s ease, opacity 0.15s;
		letter-spacing: -0.01em;
	}

	.cta-btn:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 36px color-mix(in srgb, var(--color-symon) 52%, transparent);
	}

	.cta-btn:active { transform: translateY(0); opacity: 0.88; }
	.cta-btn:disabled { opacity: 0.4; cursor: default; transform: none; }

	/* ── Content sections ─────────────────────────────────────────────────── */
	.section {
		position: relative;
		z-index: 1;
		max-width: 680px;
		margin: 0 auto;
		padding: var(--space-16) var(--space-6);
		border-top: 1px solid color-mix(in srgb, var(--color-symon) 20%, transparent);
	}

	.section-alt {
		background: color-mix(in srgb, var(--color-symon-light) 30%, transparent);
		border-radius: var(--radius-lg);
		border-top: none;
		margin: 0 auto var(--space-4);
	}

	.section-pushback {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
	}

	.section h2 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: clamp(var(--text-2xl), 4vw, var(--text-3xl));
		font-weight: var(--weight-bold);
		letter-spacing: -0.03em;
		line-height: 1.2;
		margin: 0 0 var(--space-4);
		color: var(--color-text);
	}

	.section p {
		color: var(--color-text-muted);
		line-height: 1.7;
		font-size: var(--text-base);
		margin: 0;
	}

	.section-blob {
		float: left;
		margin: 0 var(--space-6) var(--space-3) 0;
		opacity: 0.90;
		filter: drop-shadow(0 4px 16px color-mix(in srgb, var(--color-symon) 35%, transparent));
	}

	.section-blob-right {
		float: right;
		margin: 0 0 var(--space-3) var(--space-6);
	}

	.pushback-pair {
		margin-bottom: var(--space-6);
		filter: drop-shadow(0 6px 20px color-mix(in srgb, var(--color-symon-deep) 40%, transparent));
	}

	/* ── Footer CTA ───────────────────────────────────────────────────────── */
	.footer-cta {
		position: relative;
		z-index: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		padding: var(--space-16) var(--space-6);
		border-top: 1px solid color-mix(in srgb, var(--color-symon) 20%, transparent);
		gap: var(--space-4);
	}

	.footer-blob {
		filter: drop-shadow(0 8px 28px color-mix(in srgb, var(--color-symon) 40%, transparent));
		animation: hero-breathe 4s ease-in-out infinite;
	}

	@media (prefers-reduced-motion: reduce) {
		.footer-blob { animation: none; }
	}

	.footer-cta h2 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-2xl);
		font-weight: var(--weight-bold);
		letter-spacing: -0.03em;
		margin: 0;
		color: var(--color-text);
	}

	.footer-cta p {
		color: var(--color-text-muted);
		margin: 0;
		font-size: var(--text-base);
	}

	.email-row {
		display: flex;
		gap: var(--space-3);
		flex-wrap: wrap;
		justify-content: center;
		width: 100%;
		max-width: 480px;
	}

	/* ── Email input (shared) ─────────────────────────────────────────────── */
	.email-input {
		flex: 1;
		min-width: 220px;
		background: var(--color-surface);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-full);
		color: var(--color-text);
		font-family: var(--font-body);
		font-size: var(--text-base);
		padding: var(--space-4) var(--space-6);
		outline: none;
		transition: border-color 0.15s;
	}

	.email-input:focus {
		border-color: var(--color-symon);
		box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-symon) 18%, transparent);
	}

	/* ── Reveal overlay ───────────────────────────────────────────────────── */
	.reveal-backdrop {
		position: fixed;
		inset: 0;
		background: color-mix(in srgb, var(--color-symon-subtle) 96%, transparent);
		backdrop-filter: blur(8px);
		z-index: 100;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-6);
		overflow-y: auto;
		animation: fade-in 0.4s ease;
	}

	@keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }

	.reveal-esc {
		position: fixed;
		top: var(--space-4);
		right: var(--space-6);
		background: none;
		border: none;
		color: var(--color-border);
		font-size: 1.5rem;
		cursor: pointer;
		padding: var(--space-2);
		line-height: 1;
		z-index: 101;
		transition: color 0.15s;
	}
	.reveal-esc:hover { color: var(--color-text-subtle); }

	.reveal-inner {
		max-width: 560px;
		width: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-8);
		text-align: center;
	}

	.reveal-symon-img {
		filter: drop-shadow(0 8px 28px color-mix(in srgb, var(--color-symon) 45%, transparent));
		animation: hero-breathe 3s ease-in-out infinite;
	}

	/* ── Dossier ──────────────────────────────────────────────────────────── */
	.dossier { width: 100%; }

	.dossier-title {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-xl);
		font-weight: var(--weight-bold);
		letter-spacing: -0.02em;
		margin: 0 0 var(--space-4);
		color: var(--color-text);
	}

	.dossier-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 2px; }

	.dossier-item {
		display: flex;
		justify-content: space-between;
		padding: var(--space-3) var(--space-4);
		background: var(--color-symon-light);
		border-radius: var(--radius-md);
		animation: slide-in 0.3s ease;
		font-size: var(--text-sm);
	}

	@keyframes slide-in { from { opacity: 0; transform: translateX(-8px); } to { opacity: 1; transform: translateX(0); } }

	.dossier-label { color: var(--color-text-muted); }
	.dossier-value { color: var(--color-text); font-family: var(--font-mono); font-size: var(--text-xs); }

	/* ── Symon speech ─────────────────────────────────────────────────────── */
	.symon-speech { max-width: 400px; animation: fade-in 0.6s ease; }

	.symon-speech p {
		color: var(--color-text-muted);
		font-size: var(--text-lg);
		line-height: 1.6;
		margin: 0 0 var(--space-3);
	}

	.symon-punchline {
		color: var(--color-text) !important;
		font-size: var(--text-xl) !important;
		margin-top: var(--space-4) !important;
	}

	.symon-punchline em { color: var(--color-symon-deep); font-style: italic; }

	/* ── Pitch ────────────────────────────────────────────────────────────── */
	.pitch { max-width: 500px; animation: fade-in 0.6s ease; }

	.pitch h2 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-2xl);
		font-weight: var(--weight-bold);
		letter-spacing: -0.03em;
		margin: 0 0 var(--space-4);
		color: var(--color-text);
	}

	.pitch p { color: var(--color-text-muted); line-height: 1.65; margin: 0 0 var(--space-4); }

	.pitch-cta {
		font-size: var(--text-lg) !important;
		color: var(--color-text) !important;
		font-weight: var(--weight-semibold);
		margin: var(--space-6) 0 !important;
	}

	/* ── WTP ──────────────────────────────────────────────────────────────── */
	.wtp { max-width: 460px; animation: fade-in 0.6s ease; }

	.wtp h2 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-2xl);
		font-weight: var(--weight-bold);
		letter-spacing: -0.03em;
		margin: 0 0 var(--space-3);
		color: var(--color-text);
	}

	.wtp p { color: var(--color-text-muted); line-height: 1.65; margin: 0 0 var(--space-6); }

	.wtp-input-row { display: flex; align-items: center; justify-content: center; gap: var(--space-3); margin-bottom: var(--space-4); }
	.wtp-symbol, .wtp-period { color: var(--color-text-subtle); font-size: var(--text-lg); }

	.wtp-input {
		background: var(--color-surface);
		border: 1.5px solid var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text);
		font-size: 2rem;
		font-family: var(--font-display);
		width: 130px;
		text-align: center;
		padding: var(--space-3);
		outline: none;
		transition: border-color 0.15s;
	}
	.wtp-input:focus { border-color: var(--color-symon); }
	.wtp-input::-webkit-inner-spin-button, .wtp-input::-webkit-outer-spin-button { -webkit-appearance: none; }

	.wtp-thanks { color: var(--color-success); font-size: var(--text-lg); }

	/* ── Done ─────────────────────────────────────────────────────────────── */
	.done { max-width: 420px; animation: fade-in 0.6s ease; display: flex; flex-direction: column; align-items: center; gap: var(--space-4); }
	.done-blob { filter: drop-shadow(0 6px 20px color-mix(in srgb, var(--color-symon) 40%, transparent)); }

	.done h2 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-2xl);
		font-weight: var(--weight-bold);
		letter-spacing: -0.03em;
		margin: 0;
		color: var(--color-text);
	}

	.done p { color: var(--color-text-muted); margin: 0; }
</style>
