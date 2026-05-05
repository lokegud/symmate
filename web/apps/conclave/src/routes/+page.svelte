<script lang="ts">
	import { onMount } from 'svelte';
	import Box from '$lib/Box.svelte';
	import AgentHub from '$lib/modules/AgentHub.svelte';
	import Chat from '$lib/modules/Chat.svelte';
	import Canvas from '$lib/modules/Canvas.svelte';
	import { boxStore, agentStore, HUB_URL, type ModuleType, type BoxState } from '$lib/store.svelte.js';

	// SSE connection to hub server
	onMount(() => {
		let es: EventSource;
		let retryTimer: ReturnType<typeof setTimeout>;

		function connect() {
			es = new EventSource(`${HUB_URL}/events`);
			es.onmessage = (e) => {
				try {
					const msg = JSON.parse(e.data);
					if (msg.type === 'agents') agentStore.set(msg.agents ?? []);
				} catch {}
			};
			es.onerror = () => {
				es.close();
				retryTimer = setTimeout(connect, 3000);
			};
		}
		connect();
		return () => { es?.close(); clearTimeout(retryTimer); };
	});

	// Add a new box
	let adding = $state(false);
	let newModule = $state<ModuleType>('hub');

	function addBox() {
		const id = `box-${Date.now()}`;
		const colors: Record<ModuleType, string> = { hub: '#3d8ef0', chat: '#8b5cf6', canvas: '#10b981' };
		const titles: Record<ModuleType, string> = { hub: 'Agents', chat: 'Chat', canvas: 'Canvas' };
		boxStore.add({
			id,
			module: newModule,
			title: titles[newModule],
			tint: colors[newModule],
			x: 80 + Math.random() * 100,
			y: 80 + Math.random() * 100,
			w: newModule === 'chat' ? 600 : 400,
			h: newModule === 'canvas' ? 400 : 500
		});
		adding = false;
	}

	const MODULE_COMPONENTS: Record<ModuleType, any> = {
		hub: AgentHub,
		chat: Chat,
		canvas: Canvas
	};
</script>

<!-- Subtle grid pattern for depth -->
<div class="shell">
	<div class="grid-bg"></div>

	<!-- Boxes -->
	{#each boxStore.boxes as box, i (box.id)}
		<Box {box} zIndex={100 + i}>
			{#if box.module === 'hub'}
				<AgentHub />
			{:else if box.module === 'chat'}
				<Chat />
			{:else if box.module === 'canvas'}
				<Canvas />
			{/if}
		</Box>
	{/each}

	<!-- Toolbar -->
	<div class="toolbar">
		<span class="wordmark">⬡ conclave</span>

		<div class="toolbar-actions">
			{#if adding}
				<select bind:value={newModule}>
					<option value="hub">Agent Hub</option>
					<option value="chat">Chat</option>
					<option value="canvas">Canvas</option>
				</select>
				<button class="btn-add" onclick={addBox}>Add</button>
				<button class="btn-cancel" onclick={() => adding = false}>Cancel</button>
			{:else}
				<button class="btn-new" onclick={() => adding = true}>+ Box</button>
			{/if}
		</div>

		<div class="hub-status">
			<span class="dot" class:connected={agentStore.agents.length >= 0}></span>
			<span>{agentStore.agents.length} agent{agentStore.agents.length !== 1 ? 's' : ''}</span>
		</div>
	</div>
</div>

<style>
	.shell {
		width: 100vw;
		height: 100vh;
		position: relative;
		overflow: hidden;
		background: rgba(8, 8, 14, 0.92);
	}

	.grid-bg {
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(rgba(255,255,255,0.015) 1px, transparent 1px),
			linear-gradient(90deg, rgba(255,255,255,0.015) 1px, transparent 1px);
		background-size: 48px 48px;
		pointer-events: none;
	}

	.toolbar {
		position: fixed;
		bottom: 16px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		align-items: center;
		gap: 16px;
		background: rgba(12, 12, 20, 0.85);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255,255,255,0.08);
		border-radius: 40px;
		padding: 8px 20px;
		z-index: 9999;
		box-shadow: 0 8px 32px rgba(0,0,0,0.5);
	}

	.wordmark {
		font-family: ui-monospace, monospace;
		font-size: 12px;
		font-weight: 600;
		color: rgba(255,255,255,0.35);
		letter-spacing: 0.08em;
	}

	.toolbar-actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.btn-new {
		background: rgba(255,255,255,0.07);
		border: 1px solid rgba(255,255,255,0.1);
		border-radius: 20px;
		color: rgba(255,255,255,0.6);
		font-size: 12px;
		font-family: ui-monospace, monospace;
		padding: 4px 14px;
		cursor: pointer;
		transition: all 0.15s;
	}
	.btn-new:hover { background: rgba(255,255,255,0.12); color: rgba(255,255,255,0.9); }

	select {
		background: rgba(255,255,255,0.07);
		border: 1px solid rgba(255,255,255,0.1);
		border-radius: 6px;
		color: rgba(255,255,255,0.7);
		font-size: 12px;
		font-family: ui-monospace, monospace;
		padding: 4px 8px;
	}

	.btn-add {
		background: rgba(61, 142, 240, 0.2);
		border: 1px solid rgba(61, 142, 240, 0.35);
		border-radius: 20px;
		color: rgba(255,255,255,0.8);
		font-size: 12px;
		font-family: ui-monospace, monospace;
		padding: 4px 14px;
		cursor: pointer;
	}

	.btn-cancel {
		background: none;
		border: none;
		color: rgba(255,255,255,0.3);
		font-size: 12px;
		cursor: pointer;
	}

	.hub-status {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 11px;
		font-family: ui-monospace, monospace;
		color: rgba(255,255,255,0.3);
	}

	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: rgba(255,255,255,0.15);
	}
	.dot.connected { background: #10b981; }
</style>
