<script lang="ts">
	interface Message {
		role: 'user' | 'assistant';
		content: string;
	}

	let messages = $state<Message[]>([]);
	let input = $state('');
	let streaming = $state(false);
	let errorMsg = $state('');
	let bottomEl: HTMLElement;
	let textareaEl: HTMLTextAreaElement;

	function scrollToBottom() {
		bottomEl?.scrollIntoView({ behavior: 'smooth', block: 'end' });
	}

	function resizeTextarea() {
		if (!textareaEl) return;
		textareaEl.style.height = 'auto';
		textareaEl.style.height = Math.min(textareaEl.scrollHeight, 120) + 'px';
	}

	async function send() {
		const text = input.trim();
		if (!text || streaming) return;

		input = '';
		errorMsg = '';
		if (textareaEl) textareaEl.style.height = 'auto';
		messages = [...messages, { role: 'user', content: text }];
		messages = [...messages, { role: 'assistant', content: '' }];
		streaming = true;

		try {
			const res = await fetch('/chat', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ message: text })
			});

			if (!res.ok || !res.body) {
				throw new Error(`${res.status} ${res.statusText}`);
			}

			const reader = res.body.getReader();
			const decoder = new TextDecoder();
			let buf = '';

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buf += decoder.decode(value, { stream: true });
				const lines = buf.split('\n');
				buf = lines.pop() ?? '';

				for (const line of lines) {
					if (!line.startsWith('data: ')) continue;
					const token = line.slice(6);
					if (token === '[DONE]') break;

					messages[messages.length - 1] = {
						role: 'assistant',
						content: messages[messages.length - 1].content + token
					};
					scrollToBottom();
				}
			}
		} catch (e: unknown) {
			errorMsg = e instanceof Error ? e.message : 'Connection error';
			messages.pop();
		} finally {
			streaming = false;
		}
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			send();
		}
	}
</script>

<svelte:head>
	<title>Chat · Symmate</title>
</svelte:head>

<div class="chat-shell">
	<div class="topbar">
		<a href="/" class="back">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
				<path d="M19 12H5M12 5l-7 7 7 7"/>
			</svg>
			Back
		</a>
		<div class="topbar-mascot" aria-hidden="true">
			<img src="/mascot/mascot-idle.svg" alt="" width="28" height="33" class="topbar-mascot-img" />
			<span class="topbar-name">Symmate</span>
		</div>
		<div class="topbar-spacer"></div>
	</div>

	<div class="messages" role="log" aria-live="polite" aria-label="Conversation">
		{#if messages.length === 0}
			<div class="empty-state">
				<img src="/mascot/mascot-idle.svg" alt="" width="96" height="114" class="empty-mascot" aria-hidden="true"/>
				<p class="empty-greeting">Say hello</p>
				<p class="empty-notice">Conversations may be reviewed by the account owner.</p>
			</div>
		{/if}

		{#each messages as msg, i}
			<div class="message-row {msg.role}">
				{#if msg.role === 'assistant'}
					<div class="avatar" aria-hidden="true">
						<img src="/mascot/{streaming && i === messages.length - 1 ? 'mascot-listening' : 'mascot-idle'}.svg" alt="" width="28" height="33"/>
					</div>
				{/if}

				<div class="bubble">
					{#if msg.content}
						{msg.content}
					{:else if streaming && msg.role === 'assistant'}
						<span class="cursor" aria-hidden="true">▌</span>
					{/if}
				</div>

				{#if msg.role === 'user'}
					<div class="user-avatar" aria-hidden="true">
						<svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14" aria-hidden="true">
							<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7"/>
						</svg>
					</div>
				{/if}
			</div>
		{/each}

		{#if errorMsg}
			<p class="error-msg">{errorMsg}</p>
		{/if}

		<div bind:this={bottomEl} aria-hidden="true" style="height:1px"></div>
	</div>

	<form class="input-bar" onsubmit={(e) => { e.preventDefault(); send(); }}>
		<textarea
			bind:this={textareaEl}
			bind:value={input}
			onkeydown={onKeydown}
			oninput={resizeTextarea}
			placeholder="Message…"
			rows="1"
			disabled={streaming}
			aria-label="Message input"
		></textarea>
		<button type="submit" disabled={streaming || !input.trim()} aria-label="Send">
			{#if streaming}
				<span class="send-spinner" aria-hidden="true"></span>
			{:else}
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" width="16" height="16" aria-hidden="true">
					<path d="M12 19V5M5 12l7-7 7 7"/>
				</svg>
			{/if}
		</button>
	</form>
</div>

<style>
	:global(html, body) {
		height: 100%;
		overflow: hidden;
		background: var(--color-bg);
		color: var(--color-text);
		font-family: var(--font-body);
	}

	.chat-shell {
		height: 100dvh;
		display: flex;
		flex-direction: column;
		max-width: 640px;
		margin: 0 auto;
	}

	/* Top bar */
	.topbar {
		display: flex;
		align-items: center;
		padding: var(--space-3) var(--space-4);
		border-bottom: 1px solid var(--color-border-subtle);
		flex-shrink: 0;
		gap: var(--space-2);
	}

	.back {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--color-text-muted);
		text-decoration: none;
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		min-height: 44px;
		padding: var(--space-2) var(--space-1);
		transition: color 0.12s;
	}

	.back:hover { color: var(--color-text); }

	.topbar-mascot {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
	}

	.topbar-mascot-img {
		object-fit: contain;
	}

	.topbar-name {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		letter-spacing: -0.02em;
	}

	.topbar-spacer { width: 56px; }

	/* Messages */
	.messages {
		flex: 1;
		overflow-y: auto;
		padding: var(--space-4) var(--space-4) var(--space-2);
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		-webkit-overflow-scrolling: touch;
	}

	/* Empty state */
	.empty-state {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: var(--space-3);
		padding: var(--space-8) 0;
		color: var(--color-text-subtle);
		text-align: center;
	}

	.empty-mascot {
		object-fit: contain;
		animation: mascot-breathe 3s ease-in-out infinite;
		filter: drop-shadow(0 6px 18px color-mix(in srgb, var(--color-accent) 25%, transparent));
	}

	@keyframes mascot-breathe {
		0%, 100% { transform: translateY(0) scale(1); }
		50%       { transform: translateY(-5px) scale(1.02); }
	}

	@media (prefers-reduced-motion: reduce) {
		.empty-mascot { animation: none; }
	}

	.empty-greeting {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-xl);
		font-weight: var(--weight-bold);
		color: var(--color-text-muted);
		letter-spacing: -0.02em;
	}

	.empty-notice {
		font-size: var(--text-xs);
		color: var(--color-text-subtle);
		max-width: 220px;
		line-height: 1.5;
	}

	/* Message rows */
	.message-row {
		display: flex;
		align-items: flex-end;
		gap: var(--space-2);
		max-width: 85%;
	}

	.message-row.user {
		align-self: flex-end;
		flex-direction: row-reverse;
	}

	.message-row.assistant {
		align-self: flex-start;
	}

	/* Avatars */
	.avatar {
		flex-shrink: 0;
		width: 28px;
		display: flex;
		align-items: flex-end;
	}

	.avatar img {
		width: 28px;
		height: 33px;
		object-fit: contain;
	}

	.user-avatar {
		flex-shrink: 0;
		width: 28px;
		height: 28px;
		border-radius: var(--radius-full);
		background: var(--color-accent-subtle);
		color: var(--color-accent);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	/* Bubbles */
	.bubble {
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-lg);
		line-height: 1.55;
		font-size: var(--text-base);
		white-space: pre-wrap;
		word-break: break-word;
	}

	.message-row.user .bubble {
		background: var(--color-accent);
		color: #fff;
		border-bottom-right-radius: var(--radius-sm);
	}

	.message-row.assistant .bubble {
		background: var(--color-surface-raised);
		color: var(--color-text);
		border: 1px solid var(--color-border-subtle);
		border-bottom-left-radius: var(--radius-sm);
	}

	.cursor {
		display: inline-block;
		opacity: 1;
		animation: blink 1s step-end infinite;
	}

	@keyframes blink {
		0%, 100% { opacity: 1; }
		50%       { opacity: 0; }
	}

	/* Error */
	.error-msg {
		align-self: center;
		font-size: var(--text-sm);
		color: var(--color-error);
		background: var(--color-surface-raised);
		border: 1px solid color-mix(in srgb, var(--color-error) 25%, transparent);
		border-radius: var(--radius-md);
		padding: var(--space-2) var(--space-4);
	}

	/* Input bar */
	.input-bar {
		display: flex;
		align-items: flex-end;
		gap: var(--space-2);
		padding: var(--space-3) var(--space-4);
		padding-bottom: max(var(--space-3), env(safe-area-inset-bottom));
		border-top: 1px solid var(--color-border-subtle);
		background: var(--color-bg);
		flex-shrink: 0;
	}

	textarea {
		flex: 1;
		background: var(--color-surface-raised);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		color: var(--color-text);
		padding: var(--space-3) var(--space-4);
		font-size: var(--text-base);
		font-family: var(--font-body);
		resize: none;
		line-height: 1.5;
		min-height: 44px;
		max-height: 120px;
		overflow-y: auto;
		outline: none;
		transition: border-color 0.15s;
	}

	textarea::placeholder { color: var(--color-text-subtle); }

	textarea:focus {
		border-color: var(--color-accent);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-accent) 18%, transparent);
	}

	textarea:disabled { opacity: 0.5; }

	button {
		flex-shrink: 0;
		width: 44px;
		height: 44px;
		border-radius: var(--radius-full);
		background: var(--color-accent);
		color: #fff;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: opacity 0.15s, transform 0.1s;
	}

	button:disabled {
		opacity: 0.35;
		cursor: default;
	}

	button:not(:disabled):hover {
		opacity: 0.88;
		transform: scale(1.05);
	}

	button:not(:disabled):active {
		transform: scale(0.97);
	}

	/* Streaming spinner */
	.send-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255,255,255,0.35);
		border-top-color: #fff;
		border-radius: var(--radius-full);
		animation: spin 0.7s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}
</style>
