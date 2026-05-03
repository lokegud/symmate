<svelte:head>
	<title>Symmate — Your agent on the right side of A2A</title>
	<meta name="description" content="Privacy-first AI representative. One per person, one per business. Speaks for you, protects you, and has standing to say no." />
</svelte:head>

<nav>
	<span class="wordmark">Symmate</span>
	<a href="https://loke.symmate.com" class="nav-link">See a live Symmate →</a>
</nav>

<main>
	<!-- Hero -->
	<section class="hero">
		<div class="hero-inner">
			<p class="eyebrow">Agent sovereignty for everyone</p>
			<h1>Your agent on the right<br />side of A2A</h1>
			<p class="lead">
				150 enterprise agents, all built for companies talking to companies.
				None built for the person on the other side of that conversation.
				Symmate is yours.
			</p>
			<div class="hero-ctas">
				<a href="#what" class="btn-primary">What is a Symmate?</a>
				<a href="https://loke.symmate.com" class="btn-ghost">See Loke's live →</a>
			</div>
		</div>
		<div class="blob-wrap" aria-hidden="true">
			<div class="blob"></div>
		</div>
	</section>

	<!-- Gap -->
	<section class="gap-section">
		<div class="gap-inner">
			<div class="stat">
				<span class="stat-num">150</span>
				<span class="stat-label">companies on A2A<br /><em>all of them enterprise-to-enterprise</em></span>
			</div>
			<div class="gap-arrow" aria-hidden="true">→</div>
			<div class="stat highlight">
				<span class="stat-num">0</span>
				<span class="stat-label">built to represent<br /><em>you</em></span>
			</div>
		</div>
		<p class="gap-copy">
			Google's agent talks to Salesforce's agent. Microsoft's Copilot talks to ServiceNow.
			The protocols are real, the volumes are real, the money is real.
			The slot that's missing is the one that argues for you.
		</p>
	</section>

	<!-- What is a Symmate -->
	<section id="what" class="what-section">
		<h2>What is a Symmate?</h2>
		<p class="section-sub">
			An AI agent that represents you — not as a chatbot on your website, but as your
			<em>agent endpoint</em> in a world where AI talks to AI.
		</p>

		<div class="cards">
			<div class="card">
				<div class="card-icon" aria-hidden="true">🎯</div>
				<h3>Represents you</h3>
				<p>Speaks for you to other agents, visitors, and services. Knows your scope, your values, and what you'd actually say.</p>
			</div>
			<div class="card">
				<div class="card-icon" aria-hidden="true">🔒</div>
				<h3>Privacy firewall built in</h3>
				<p>Every outbound response goes through a capability policy you control. PII only leaves when you say so. Audit log of every decision.</p>
			</div>
			<div class="card">
				<div class="card-icon" aria-hidden="true">✋</div>
				<h3>Has standing to say no</h3>
				<p>Your Symmate knows your values well enough to push back. Not a yes-machine — someone who represents you, which includes the occasional refusal.</p>
			</div>
			<div class="card">
				<div class="card-icon" aria-hidden="true">🌱</div>
				<h3>Grows with you</h3>
				<p>Trust accrues through shared history. Symmate learns your patterns, your edge cases, your exceptions — and earns the right to handle them.</p>
			</div>
		</div>
	</section>

	<!-- How it works -->
	<section class="how-section">
		<h2>How it works</h2>
		<ol class="steps">
			<li>
				<span class="step-num">1</span>
				<div>
					<strong>Snap-to-personality</strong>
					<p>A quick onboarding builds your Symmate's voice, values, and scope from how you actually communicate — not from a personality quiz.</p>
				</div>
			</li>
			<li>
				<span class="step-num">2</span>
				<div>
					<strong>Your Symmate goes live</strong>
					<p>Visitors can chat with it. Other agents can query it via A2A. It handles what it can and escalates what it can't.</p>
				</div>
			</li>
			<li>
				<span class="step-num">3</span>
				<div>
					<strong>You stay in control</strong>
					<p>Owner dashboard to edit personality, review conversations, update policy, and flip a kill switch if needed. No magic black box.</p>
				</div>
			</li>
		</ol>
	</section>

	<!-- Footer CTA -->
	<section class="cta-section">
		<h2>First Symmate is Loke's</h2>
		<p>We built it for ourselves first. <a href="https://loke.symmate.com">loke.symmate.com</a> is running now, talking to real people and other agents over Yggdrasil and standard A2A.</p>
		<p class="cta-sub">Interested in being next? <a href="mailto:loke@justadestination.com">Say hello →</a></p>
	</section>
</main>

<footer>
	<div class="footer-inner">
		<span class="wordmark-sm">Symmate</span>
		<div class="footer-links">
			<a href="/.well-known/agent-card.json">agent card</a>
			<a href="/llms.txt">llms.txt</a>
			<a href="https://loke.symmate.com">live demo</a>
		</div>
		<p class="footer-copy">
			Built by Loke &amp; Mox. Open source.
			Privacy-first by design, not by policy.
		</p>
	</div>
</footer>

<style>
	:global(html) {
		scroll-behavior: smooth;
	}
	:global(body) {
		background: var(--color-bg);
		color: var(--color-text);
		font-family: var(--font-body);
		margin: 0;
		min-height: 100dvh;
		line-height: 1.6;
	}

	/* Nav */
	nav {
		position: sticky;
		top: 0;
		z-index: 10;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 2rem;
		background: var(--color-surface-overlay, rgba(14,12,10,0.85));
		backdrop-filter: blur(12px);
		border-bottom: 1px solid var(--color-border-subtle);
	}
	.wordmark {
		font-family: var(--font-display);
		font-size: 1.25rem;
		font-weight: 600;
		letter-spacing: -0.02em;
		color: var(--color-text);
	}
	.nav-link {
		font-size: var(--text-sm);
		color: var(--color-accent);
		text-decoration: none;
	}
	.nav-link:hover { text-decoration: underline; }

	/* Hero */
	.hero {
		position: relative;
		overflow: hidden;
		padding: 6rem 2rem 5rem;
		max-width: 72rem;
		margin: 0 auto;
		display: grid;
		grid-template-columns: 1fr;
		gap: 3rem;
		align-items: center;
	}
	@media (min-width: 768px) {
		.hero { grid-template-columns: 1fr 1fr; padding: 8rem 3rem 6rem; }
	}
	.hero-inner { max-width: 560px; }
	.eyebrow {
		font-size: var(--text-sm);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--color-accent);
		margin: 0 0 1rem;
	}
	h1 {
		font-family: var(--font-display);
		font-size: clamp(2.25rem, 5vw, 4rem);
		font-weight: 700;
		line-height: 1.1;
		letter-spacing: -0.03em;
		margin: 0 0 1.5rem;
		color: var(--color-text);
	}
	.lead {
		font-size: var(--text-lg);
		color: var(--color-text-muted);
		margin: 0 0 2.5rem;
		line-height: 1.65;
	}
	.hero-ctas { display: flex; gap: 1rem; flex-wrap: wrap; }
	.btn-primary {
		background: var(--color-accent);
		color: var(--color-accent-text, #fff);
		padding: 0.85rem 1.75rem;
		border-radius: 8px;
		font-weight: 600;
		text-decoration: none;
		font-size: var(--text-base);
		transition: background 0.15s;
	}
	.btn-primary:hover { background: var(--color-accent-hover); }
	.btn-ghost {
		color: var(--color-text-muted);
		padding: 0.85rem 0;
		text-decoration: none;
		font-size: var(--text-base);
		border-bottom: 1px solid var(--color-border);
	}
	.btn-ghost:hover { color: var(--color-text); border-color: var(--color-text-muted); }

	/* Blob */
	.blob-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.blob {
		width: min(280px, 70vw);
		height: min(280px, 70vw);
		background: conic-gradient(from 0deg, #4B5FD3, #7C6FE8, #A06EE0, #6C9FD3, #4B5FD3);
		border-radius: 60% 40% 55% 45% / 50% 45% 55% 50%;
		animation: blob-breathe 8s ease-in-out infinite, prism-shift 6s linear infinite;
		background-size: 200% 200%;
		opacity: 0.85;
		filter: blur(1px);
	}
	@keyframes blob-breathe {
		0%   { border-radius: 60% 40% 55% 45% / 50% 45% 55% 50%; }
		25%  { border-radius: 50% 50% 40% 60% / 55% 40% 60% 45%; }
		50%  { border-radius: 40% 60% 60% 40% / 45% 55% 45% 55%; }
		75%  { border-radius: 55% 45% 45% 55% / 40% 60% 50% 50%; }
		100% { border-radius: 60% 40% 55% 45% / 50% 45% 55% 50%; }
	}
	@keyframes prism-shift {
		0%   { filter: hue-rotate(0deg) blur(1px); }
		100% { filter: hue-rotate(360deg) blur(1px); }
	}

	/* Gap section */
	.gap-section {
		background: var(--color-surface);
		border-top: 1px solid var(--color-border-subtle);
		border-bottom: 1px solid var(--color-border-subtle);
		padding: 5rem 2rem;
		text-align: center;
	}
	.gap-inner {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 3rem;
		flex-wrap: wrap;
		margin-bottom: 2rem;
	}
	.stat { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; }
	.stat-num {
		font-family: var(--font-display);
		font-size: clamp(3rem, 8vw, 5.5rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		color: var(--color-text-subtle);
		line-height: 1;
	}
	.stat.highlight .stat-num { color: var(--color-accent); }
	.stat-label {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		text-align: center;
	}
	.stat-label em { color: var(--color-text-subtle); font-style: normal; }
	.gap-arrow {
		font-size: 2rem;
		color: var(--color-border-strong, var(--color-border));
	}
	.gap-copy {
		max-width: 640px;
		margin: 0 auto;
		color: var(--color-text-muted);
		line-height: 1.7;
		font-size: var(--text-base);
	}

	/* What section */
	.what-section {
		padding: 6rem 2rem;
		max-width: 72rem;
		margin: 0 auto;
	}
	h2 {
		font-family: var(--font-display);
		font-size: clamp(1.875rem, 4vw, 3rem);
		font-weight: 700;
		letter-spacing: -0.03em;
		margin: 0 0 1rem;
		color: var(--color-text);
	}
	.section-sub {
		font-size: var(--text-lg);
		color: var(--color-text-muted);
		max-width: 600px;
		margin: 0 0 3.5rem;
		line-height: 1.65;
	}
	.section-sub em { color: var(--color-text); font-style: italic; }
	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 1.5rem;
	}
	.card {
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: 12px;
		padding: 1.75rem;
	}
	.card-icon { font-size: 1.75rem; margin-bottom: 1rem; }
	.card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		margin: 0 0 0.6rem;
		color: var(--color-text);
		font-family: var(--font-display);
	}
	.card p { margin: 0; color: var(--color-text-muted); font-size: var(--text-sm); line-height: 1.65; }

	/* How section */
	.how-section {
		background: var(--color-surface);
		border-top: 1px solid var(--color-border-subtle);
		padding: 6rem 2rem;
	}
	.how-section h2 { max-width: 72rem; margin: 0 auto 3rem; }
	.steps {
		list-style: none;
		padding: 0;
		margin: 0 auto;
		max-width: 640px;
		display: flex;
		flex-direction: column;
		gap: 2.5rem;
	}
	.steps li { display: flex; gap: 1.5rem; align-items: flex-start; }
	.step-num {
		flex-shrink: 0;
		width: 2.25rem;
		height: 2.25rem;
		border-radius: 50%;
		background: var(--color-accent-subtle);
		color: var(--color-accent);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: var(--text-sm);
		margin-top: 0.15rem;
	}
	.steps strong { display: block; font-size: var(--text-md); margin-bottom: 0.4rem; color: var(--color-text); }
	.steps p { margin: 0; color: var(--color-text-muted); font-size: var(--text-sm); line-height: 1.65; }

	/* CTA section */
	.cta-section {
		padding: 6rem 2rem;
		max-width: 640px;
		margin: 0 auto;
		text-align: center;
	}
	.cta-section h2 { margin: 0 0 1.25rem; }
	.cta-section p { color: var(--color-text-muted); margin: 0 0 1rem; line-height: 1.7; }
	.cta-section a { color: var(--color-accent); }
	.cta-sub { font-size: var(--text-lg) !important; color: var(--color-text) !important; }

	/* Footer */
	footer {
		border-top: 1px solid var(--color-border-subtle);
		padding: 2.5rem 2rem;
	}
	.footer-inner {
		max-width: 72rem;
		margin: 0 auto;
		display: flex;
		flex-wrap: wrap;
		gap: 1.5rem;
		align-items: center;
	}
	.wordmark-sm {
		font-family: var(--font-display);
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text-subtle);
	}
	.footer-links {
		display: flex;
		gap: 1.5rem;
		flex: 1;
	}
	.footer-links a {
		font-size: var(--text-sm);
		color: var(--color-text-subtle);
		text-decoration: none;
	}
	.footer-links a:hover { color: var(--color-text-muted); }
	.footer-copy {
		font-size: var(--text-xs);
		color: var(--color-text-subtle);
		margin: 0;
		width: 100%;
	}
</style>
