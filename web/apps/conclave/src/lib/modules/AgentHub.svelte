<script lang="ts">
	import { agentStore, HUB_URL } from '../store.svelte.js';

	const STATUS_COLOR: Record<string, string> = {
		idle:    '#6b7280',
		working: '#10b981',
		blocked: '#f59e0b',
		done:    '#3d8ef0',
	};

	let selectedAgent = $state<string | null>(null);

	async function override(agentId: string, pokerSessId: string) {
		await fetch(`${HUB_URL}/agents/${agentId}/poker-override`, { method: 'POST' });
	}
</script>

<div class="hub">
	{#if agentStore.agents.length === 0}
		<div class="empty">No agents registered</div>
	{:else}
		{#each agentStore.agents as agent (agent.id)}
			<div
				class="agent"
				class:selected={selectedAgent === agent.id}
				onclick={() => selectedAgent = selectedAgent === agent.id ? null : agent.id}
				role="button"
				tabindex="0"
				onkeydown={(e) => e.key === 'Enter' && (selectedAgent = agent.id)}
			>
				<div class="agent-header">
					<span class="status-dot" style="background: {STATUS_COLOR[agent.status]}"></span>
					<span class="agent-name">{agent.name}</span>
					<span class="status-label">{agent.status}</span>
				</div>

				<div class="agent-task">{agent.task || 'No task'}</div>

				{#if selectedAgent === agent.id && agent.todos?.length > 0}
					<div class="todos">
						{#each agent.todos as todo (todo.id)}
							<div class="todo" class:done={todo.done}>
								<span class="todo-check">{todo.done ? '✓' : '○'}</span>
								<span>{todo.text}</span>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/each}
	{/if}
</div>

<style>
	.hub {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 8px;
		height: 100%;
		overflow-y: auto;
	}

	.empty {
		color: rgba(255,255,255,0.25);
		font-size: 12px;
		text-align: center;
		padding: 32px 0;
		font-family: ui-monospace, monospace;
	}

	.agent {
		padding: 10px 12px;
		border-radius: 6px;
		cursor: pointer;
		background: rgba(255,255,255,0.03);
		border: 1px solid rgba(255,255,255,0.05);
		transition: background 0.15s;
	}
	.agent:hover { background: rgba(255,255,255,0.06); }
	.agent.selected { background: rgba(255,255,255,0.07); border-color: rgba(255,255,255,0.1); }

	.agent-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 4px;
	}

	.status-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.agent-name {
		font-family: ui-monospace, monospace;
		font-size: 12px;
		font-weight: 600;
		color: rgba(255,255,255,0.85);
		flex: 1;
	}

	.status-label {
		font-size: 10px;
		font-family: ui-monospace, monospace;
		color: rgba(255,255,255,0.35);
		letter-spacing: 0.05em;
	}

	.agent-task {
		font-size: 11px;
		color: rgba(255,255,255,0.45);
		font-family: ui-monospace, monospace;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		padding-left: 15px;
	}

	.todos {
		margin-top: 8px;
		padding-left: 15px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.todo {
		display: flex;
		gap: 8px;
		font-size: 11px;
		font-family: ui-monospace, monospace;
		color: rgba(255,255,255,0.6);
	}
	.todo.done { color: rgba(255,255,255,0.25); text-decoration: line-through; }

	.todo-check { color: #10b981; flex-shrink: 0; }
	.todo.done .todo-check { color: rgba(255,255,255,0.2); }
</style>
