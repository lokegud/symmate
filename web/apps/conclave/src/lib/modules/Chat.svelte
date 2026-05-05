<script lang="ts">
	import { agentStore, HUB_URL } from '../store.svelte.js';

	interface Message {
		role: 'user' | 'assistant';
		content: string;
		agentId?: string;
	}

	let messages = $state<Message[]>([]);
	let input = $state('');
	let streaming = $state(false);
	let selectedAgentId = $state<string>('');
	let bottomEl: HTMLElement;

	$effect(() => {
		if (agentStore.agents.length > 0 && !selectedAgentId) {
			selectedAgentId = agentStore.agents[0].id;
		}
	});

	$effect(() => {
		bottomEl?.scrollIntoView({ behavior: 'smooth', block: 'end' });
	});

	async function send() {
		const text = input.trim();
		if (!text || streaming || !selectedAgentId) return;
		input = '';

		messages = [...messages, { role: 'user', content: text }];
		messages = [...messages, { role: 'assistant', content: '', agentId: selectedAgentId }];
		streaming = true;

		try {
			const res = await fetch(`${HUB_URL}/agents/${selectedAgentId}/chat`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ message: text })
			});

			if (!res.ok || !res.body) throw new Error(`${res.status}`);

			const reader = res.body.getReader();
			const decoder = new TextDecoder();

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;
				const chunk = decoder.decode(value, { stream: true });
				messages[messages.length - 1] = {
					...messages[messages.length - 1],
					content: messages[messages.length - 1].content + chunk
				};
			}
		} catch (err) {
			messages[messages.length - 1] = {
				...messages[messages.length - 1],
				content: `[error: ${err}]`
			};
		} finally {
			streaming = false;
		}
	}

	function onKeyDown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			send();
		}
	}

	function agentName(id: string) {
		return agentStore.agents.find(a => a.id === id)?.name ?? id;
	}
</script>

<div class="chat">
	<div class="agent-select">
		{#if agentStore.agents.length === 0}
			<span class="no-agents">No agents online</span>
		{:else}
			{#each agentStore.agents as agent (agent.id)}
				<button
					class="agent-chip"
					class:active={selectedAgentId === agent.id}
					onclick={() => selectedAgentId = agent.id}
				>
					{agent.name}
				</button>
			{/each}
		{/if}
	</div>

	<div class="messages">
		{#each messages as msg, i (i)}
			<div class="msg {msg.role}">
				{#if msg.role === 'assistant'}
					<span class="msg-from">{agentName(msg.agentId ?? '')}</span>
				{/if}
				<span class="msg-text">{msg.content}{msg.role === 'assistant' && streaming && i === messages.length - 1 ? '▋' : ''}</span>
			</div>
		{/each}
		<div bind:this={bottomEl}></div>
	</div>

	<div class="input-row">
		<textarea
			bind:value={input}
			onkeydown={onKeyDown}
			placeholder={selectedAgentId ? `Message ${agentName(selectedAgentId)}…` : 'Select an agent'}
			disabled={!selectedAgentId || streaming}
			rows="2"
		></textarea>
		<button onclick={send} disabled={!input.trim() || streaming || !selectedAgentId}>↑</button>
	</div>
</div>

<style>
	.chat {
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow: hidden;
	}

	.agent-select {
		display: flex;
		gap: 6px;
		padding: 8px 10px;
		border-bottom: 1px solid rgba(255,255,255,0.05);
		flex-wrap: wrap;
		flex-shrink: 0;
	}

	.no-agents {
		font-size: 11px;
		color: rgba(255,255,255,0.25);
		font-family: ui-monospace, monospace;
	}

	.agent-chip {
		background: rgba(255,255,255,0.06);
		border: 1px solid rgba(255,255,255,0.1);
		border-radius: 4px;
		color: rgba(255,255,255,0.5);
		font-size: 11px;
		font-family: ui-monospace, monospace;
		padding: 3px 8px;
		cursor: pointer;
		transition: all 0.15s;
	}
	.agent-chip.active {
		background: rgba(139, 92, 246, 0.2);
		border-color: rgba(139, 92, 246, 0.4);
		color: rgba(255,255,255,0.85);
	}

	.messages {
		flex: 1;
		overflow-y: auto;
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-height: 0;
	}

	.msg {
		display: flex;
		flex-direction: column;
		gap: 2px;
		max-width: 85%;
	}
	.msg.user { align-self: flex-end; align-items: flex-end; }
	.msg.assistant { align-self: flex-start; }

	.msg-from {
		font-size: 10px;
		font-family: ui-monospace, monospace;
		color: rgba(139, 92, 246, 0.6);
		letter-spacing: 0.05em;
	}

	.msg-text {
		font-size: 12px;
		font-family: ui-monospace, monospace;
		line-height: 1.55;
		padding: 8px 12px;
		border-radius: 8px;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.msg.user .msg-text {
		background: rgba(139, 92, 246, 0.18);
		border: 1px solid rgba(139, 92, 246, 0.25);
		color: rgba(255,255,255,0.85);
	}

	.msg.assistant .msg-text {
		background: rgba(255,255,255,0.05);
		border: 1px solid rgba(255,255,255,0.07);
		color: rgba(255,255,255,0.8);
	}

	.input-row {
		display: flex;
		gap: 8px;
		padding: 10px;
		border-top: 1px solid rgba(255,255,255,0.05);
		flex-shrink: 0;
	}

	textarea {
		flex: 1;
		background: rgba(255,255,255,0.05);
		border: 1px solid rgba(255,255,255,0.1);
		border-radius: 6px;
		color: rgba(255,255,255,0.85);
		font-family: ui-monospace, monospace;
		font-size: 12px;
		padding: 8px 10px;
		resize: none;
		outline: none;
		transition: border-color 0.15s;
	}
	textarea:focus { border-color: rgba(139, 92, 246, 0.4); }
	textarea::placeholder { color: rgba(255,255,255,0.2); }

	button {
		background: rgba(139, 92, 246, 0.25);
		border: 1px solid rgba(139, 92, 246, 0.35);
		border-radius: 6px;
		color: rgba(255,255,255,0.7);
		font-size: 16px;
		width: 38px;
		cursor: pointer;
		transition: all 0.15s;
		align-self: flex-end;
		height: 38px;
	}
	button:hover:not(:disabled) { background: rgba(139, 92, 246, 0.4); color: white; }
	button:disabled { opacity: 0.3; cursor: not-allowed; }
</style>
