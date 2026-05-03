<script lang="ts">
	interface Message {
		role: 'user' | 'assistant';
		content: string;
	}

	let messages = $state<Message[]>([]);
	let input = $state('');
	let streaming = $state(false);
	let tokenCount = $state(0);
	let errorMsg = $state('');
	let bottomEl: HTMLElement;

	function scrollToBottom() {
		bottomEl?.scrollIntoView({ behavior: 'smooth', block: 'end' });
	}

	async function send() {
		const text = input.trim();
		if (!text || streaming) return;

		input = '';
		errorMsg = '';
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
					tokenCount += token.length;
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
		<a href="/" class="back">← Vestibule</a>
		<span class="token-counter" title="approximate token count">{tokenCount} tokens</span>
	</div>

	<div class="messages" role="log" aria-live="polite" aria-label="Conversation">
		{#if messages.length === 0}
			<div class="empty-state">
				<p>Start a conversation.</p>
				<p class="notice">AI · responses logged for owner review</p>
			</div>
		{/if}

		{#each messages as msg}
			<div class="message {msg.role}">
				<div class="bubble">{msg.content || (streaming && msg.role === 'assistant' ? '▌' : '')}</div>
			</div>
		{/each}

		{#if errorMsg}
			<div class="error">{errorMsg}</div>
		{/if}

		<div bind:this={bottomEl} aria-hidden="true"></div>
	</div>

	<form class="input-bar" onsubmit={(e) => { e.preventDefault(); send(); }}>
		<textarea
			bind:value={input}
			onkeydown={onKeydown}
			placeholder="Send a message…"
			rows="1"
			disabled={streaming}
			aria-label="Message input"
		></textarea>
		<button type="submit" disabled={streaming || !input.trim()} aria-label="Send">
			{streaming ? '…' : '↑'}
		</button>
	</form>
</div>

<style>
	:global(*, *::before, *::after) {
		box-sizing: border-box;
		margin: 0;
		padding: 0;
	}

	:global(html, body) {
		height: 100%;
		overflow: hidden;
		font-family: system-ui, -apple-system, sans-serif;
		background: #0f0f11;
		color: #e8e8e8;
	}

	.chat-shell {
		height: 100dvh;
		display: flex;
		flex-direction: column;
	}

	.topbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #1e1e24;
		flex-shrink: 0;
	}

	.back {
		color: #6b7280;
		text-decoration: none;
		font-size: 0.875rem;
	}

	.back:hover {
		color: #9ca3af;
	}

	.token-counter {
		font-size: 0.75rem;
		color: #4b5563;
		font-variant-numeric: tabular-nums;
	}

	.messages {
		flex: 1;
		overflow-y: auto;
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.empty-state {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		color: #4b5563;
		font-size: 0.9rem;
		text-align: center;
	}

	.notice {
		font-size: 0.75rem;
		color: #374151;
	}

	.message {
		display: flex;
	}

	.message.user {
		justify-content: flex-end;
	}

	.message.assistant {
		justify-content: flex-start;
	}

	.bubble {
		max-width: 80%;
		padding: 0.65rem 0.9rem;
		border-radius: 16px;
		line-height: 1.5;
		font-size: 0.9375rem;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.message.user .bubble {
		background: #4f46e5;
		color: #fff;
		border-bottom-right-radius: 4px;
	}

	.message.assistant .bubble {
		background: #1e1e24;
		color: #e8e8e8;
		border-bottom-left-radius: 4px;
	}

	.error {
		color: #ef4444;
		font-size: 0.875rem;
		text-align: center;
		padding: 0.5rem;
	}

	.input-bar {
		display: flex;
		align-items: flex-end;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		border-top: 1px solid #1e1e24;
		flex-shrink: 0;
	}

	textarea {
		flex: 1;
		background: #1e1e24;
		border: 1px solid #2e2e38;
		border-radius: 12px;
		color: #e8e8e8;
		padding: 0.65rem 0.9rem;
		font-size: 0.9375rem;
		font-family: inherit;
		resize: none;
		line-height: 1.5;
		max-height: 120px;
		overflow-y: auto;
		outline: none;
	}

	textarea:focus {
		border-color: #4f46e5;
	}

	textarea:disabled {
		opacity: 0.5;
	}

	button {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: #4f46e5;
		color: #fff;
		border: none;
		font-size: 1.1rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		transition: opacity 0.15s;
	}

	button:disabled {
		opacity: 0.35;
		cursor: default;
	}

	button:not(:disabled):hover {
		opacity: 0.85;
	}
</style>
