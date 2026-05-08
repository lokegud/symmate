<script lang="ts">
	import { onMount } from 'svelte';
	import Box from '$lib/Box.svelte';
	import AgentHub from '$lib/modules/AgentHub.svelte';
	import Chat from '$lib/modules/Chat.svelte';
	import Canvas from '$lib/modules/Canvas.svelte';
	import CohortPresence from '$lib/modules/CohortPresence.svelte';
	import ConversationStream from '$lib/modules/ConversationStream.svelte';
	import DecisionLog from '$lib/modules/DecisionLog.svelte';
	import TaskBoard from '$lib/modules/TaskBoard.svelte';
	import SystemLog from '$lib/modules/SystemLog.svelte';
	import ResourceMeter from '$lib/modules/ResourceMeter.svelte';
	import { boxStore, agentStore, HUB_URL, type ModuleType, type Agent } from '$lib/store.svelte.js';

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

		// Mouse passthrough for Tauri HUD mode.
		// When cursor is not over a .box, clicks pass through to the desktop.
		// setIgnoreCursorEvents(true) on Linux kills ALL input to the window, so
		// re-entry detection uses mousemove (fires while events are active) rather
		// than polling a cursor-position API that doesn't exist in @tauri-apps/api 2.11.
		// Re-entry after passthrough: cursor must cross a .box boundary before events
		// re-enable — this works because mousemove still fires on the box as soon as
		// the window regains hit-test ownership on that pixel region.
		// TODO: revisit with tauri://cursor-moved event when available in a future release
		async function setupPassthrough() {
			const { getCurrentWindow } = await import('@tauri-apps/api/window');
			const win = getCurrentWindow();
			let passthrough = false;

			document.addEventListener('mousemove', async (e) => {
				const onBox = !!(e.target as HTMLElement)?.closest('.box');
				if (onBox && passthrough) {
					await win.setIgnoreCursorEvents(false);
					passthrough = false;
				} else if (!onBox && !passthrough) {
					await win.setIgnoreCursorEvents(true);
					passthrough = true;
				}
			}, { passive: true });
		}

		// Only activate in Tauri context
		if (typeof window !== 'undefined' && '__TAURI__' in window) {
			setupPassthrough();
		}

		return () => { es?.close(); clearTimeout(retryTimer); };
	});

	// Add a new box
	let adding = $state(false);
	let newModule = $state<ModuleType>('hub');

	const MODULE_META: Record<ModuleType, { title: string; tint: string; w: number; h: number }> = {
		hub:       { title: 'Agents',    tint: '#3d8ef0', w: 400, h: 500 },
		chat:      { title: 'Chat',      tint: '#8b5cf6', w: 600, h: 400 },
		canvas:    { title: 'Canvas',    tint: '#10b981', w: 500, h: 500 },
		presence:  { title: 'Presence',  tint: '#4a9eff', w: 360, h: 480 },
		stream:    { title: 'Conclave',  tint: '#e8a878', w: 560, h: 420 },
		decisions: { title: 'Decisions', tint: '#a78bfa', w: 440, h: 520 },
		taskboard: { title: 'Round',     tint: '#f59e0b', w: 540, h: 240 },
		syslog:    { title: 'Sys Log',   tint: '#6b7280', w: 520, h: 420 },
		resources: { title: 'Resources', tint: '#e59e0b', w: 300, h: 220 },
	};

	function addBox() {
		const id = `box-${Date.now()}`;
		const meta = MODULE_META[newModule];
		boxStore.add({
			id,
			module: newModule,
			title: meta.title,
			tint: meta.tint,
			x: 60 + Math.random() * 120,
			y: 60 + Math.random() * 80,
			w: meta.w,
			h: meta.h,
		});
		adding = false;
	}

	const MODULE_COMPONENTS: Record<ModuleType, any> = {
		hub:       AgentHub,
		chat:      Chat,
		canvas:    Canvas,
		presence:  CohortPresence,
		stream:    ConversationStream,
		decisions: DecisionLog,
		taskboard: TaskBoard,
		syslog:    SystemLog,
		resources: ResourceMeter,
	};
</script>

<!-- Subtle grid pattern for depth -->
<div class="shell">
	<div class="grid-bg"></div>

	<!-- Boxes -->
	{#each boxStore.boxes as box, i (box.id)}
		{@const Comp = MODULE_COMPONENTS[box.module]}
		<Box {box} zIndex={100 + i}>
			{#if Comp}<Comp />{/if}
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
					<option value="presence">Presence</option>
					<option value="stream">Conclave Stream</option>
					<option value="decisions">Decision Log</option>
					<option value="taskboard">Round Board</option>
					<option value="syslog">Sys Log</option>
					<option value="resources">Resources</option>
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
		background: var(--hub-surface-overlay);
	}

	.grid-bg {
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(var(--hub-grid-color) 1px, transparent 1px),
			linear-gradient(90deg, var(--hub-grid-color) 1px, transparent 1px);
		background-size: var(--hub-grid-size) var(--hub-grid-size);
		pointer-events: none;
	}

	/* Grid is a cockpit depth cue; meaningless over real desktop pixels */
	:global([data-hub-mode="hud"]) .grid-bg {
		display: none;
	}

	.toolbar {
		position: fixed;
		bottom: var(--hub-space-4);
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		align-items: center;
		gap: var(--hub-space-4);
		background: var(--hub-surface);
		backdrop-filter: blur(20px);
		border: 1px solid var(--hub-border);
		border-radius: var(--hub-radius-full);
		padding: var(--hub-space-2) var(--hub-space-5);
		z-index: 9999;
		box-shadow: var(--hub-shadow-box);
	}

	.wordmark {
		font-family: var(--hub-font-mono);
		font-size: var(--hub-text-xs);
		font-weight: 600;
		color: var(--hub-text-subtle);
		letter-spacing: 0.08em;
	}

	.toolbar-actions {
		display: flex;
		align-items: center;
		gap: var(--hub-space-2);
	}

	.btn-new {
		background: rgba(255,255,255,0.07);
		border: 1px solid var(--hub-border-strong);
		border-radius: var(--hub-radius-full);
		color: var(--hub-text-muted);
		font-size: var(--hub-text-xs);
		font-family: var(--hub-font-mono);
		padding: var(--hub-space-1) 14px;
		cursor: pointer;
		transition: all var(--hub-duration-fast) var(--hub-ease);
	}
	.btn-new:hover { background: rgba(255,255,255,0.12); color: var(--hub-text); }

	select {
		background: rgba(255,255,255,0.07);
		border: 1px solid var(--hub-border-strong);
		border-radius: var(--hub-radius-md);
		color: var(--hub-text-muted);
		font-size: var(--hub-text-xs);
		font-family: var(--hub-font-mono);
		padding: var(--hub-space-1) var(--hub-space-2);
	}

	.btn-add {
		background: var(--hub-accent-dim);
		border: 1px solid var(--hub-accent-border);
		border-radius: var(--hub-radius-full);
		color: var(--hub-text);
		font-size: var(--hub-text-xs);
		font-family: var(--hub-font-mono);
		padding: var(--hub-space-1) 14px;
		cursor: pointer;
		transition: all var(--hub-duration-fast) var(--hub-ease);
	}
	.btn-add:hover { background: color-mix(in srgb, var(--hub-accent) 25%, transparent); }

	.btn-cancel {
		background: none;
		border: none;
		color: var(--hub-text-subtle);
		font-size: var(--hub-text-xs);
		font-family: var(--hub-font-mono);
		cursor: pointer;
	}

	.hub-status {
		display: flex;
		align-items: center;
		gap: var(--hub-space-1);
		font-size: var(--hub-text-2xs);
		font-family: var(--hub-font-mono);
		color: var(--hub-text-subtle);
	}

	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: rgba(255,255,255,0.15);
	}
	.dot.connected { background: var(--state-working); }
</style>
