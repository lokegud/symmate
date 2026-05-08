<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const conversation = $derived(data.conversation);

	function fmtDate(ms: number) {
		return new Date(ms).toLocaleString('en-GB', {
			day: '2-digit', month: 'short', year: 'numeric',
			hour: '2-digit', minute: '2-digit', second: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>Conversation · Owner Dashboard</title>
</svelte:head>

<div class="page">
	<div class="breadcrumb">
		<a href="/owner/conversations">← Conversations</a>
	</div>

	<header class="conv-header">
		<div class="conv-meta">
			<h1 class="font-display">{conversation.visitor_handle ?? 'Anonymous'}</h1>
			<div class="conv-details">
				<span class="detail-chip">{conversation.surface}</span>
				<span class="detail-chip">{fmtDate(conversation.started_at)}</span>
				{#if conversation.outcome}
					<span class="detail-chip detail-chip--accent">{conversation.outcome}</span>
				{/if}
			</div>
		</div>
	</header>

	<div class="transcript">
		{#if conversation.messages.length === 0}
			<p class="empty-msg">No messages in this conversation.</p>
		{:else}
			{#each conversation.messages as msg}
				<div class="msg msg--{msg.role}">
					<div class="msg-role">{msg.role}</div>
					<div class="msg-body">{msg.content}</div>
					<div class="msg-ts">{fmtDate(msg.ts)}</div>
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.page {
		max-width: 800px;
		display: flex;
		flex-direction: column;
		gap: var(--space-6);
	}

	.breadcrumb a {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.breadcrumb a:hover { color: var(--color-text); }

	.conv-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.conv-header h1 {
		font-family: var(--font-display);
		font-optical-sizing: auto;
		font-size: var(--text-2xl);
		font-weight: var(--weight-semibold);
		letter-spacing: -0.03em;
		color: var(--color-text);
	}

	.conv-details {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.detail-chip {
		display: inline-flex;
		align-items: center;
		height: 24px;
		padding: 0 var(--space-3);
		background: var(--color-surface-raised);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: var(--weight-medium);
		color: var(--color-text-muted);
	}

	.detail-chip--accent {
		background: var(--color-accent-subtle);
		color: var(--color-accent);
		border-color: transparent;
	}

	.transcript {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.empty-msg {
		font-size: var(--text-sm);
		color: var(--color-text-subtle);
		text-align: center;
		padding: var(--space-12) 0;
	}

	.msg {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		padding: var(--space-4);
		border-radius: var(--radius-md);
	}

	.msg--user {
		background: var(--color-accent-subtle);
		align-self: flex-end;
		max-width: 85%;
	}

	.msg--assistant {
		background: var(--color-surface);
		border: 1px solid var(--color-border-subtle);
		align-self: flex-start;
		max-width: 85%;
	}

	.msg-role {
		font-size: var(--text-xs);
		font-weight: var(--weight-semibold);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--color-text-muted);
	}

	.msg--user .msg-role { color: var(--color-accent); }

	.msg-body {
		font-size: var(--text-sm);
		line-height: 1.6;
		color: var(--color-text);
		white-space: pre-wrap;
		word-break: break-word;
	}

	.msg-ts {
		font-size: var(--text-xs);
		font-family: var(--font-mono);
		color: var(--color-text-subtle);
		align-self: flex-end;
	}
</style>
