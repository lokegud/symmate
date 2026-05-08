<script lang="ts">
	import '$lib/styles/tokens.css';
	let { children } = $props();
</script>

{@render children()}

<style>
	:global(*, *::before, *::after) { box-sizing: border-box; margin: 0; padding: 0; }
	:global(html, body) {
		width: 100%;
		height: 100%;
		overflow: hidden;
		background: var(--hub-bg);
		color: var(--hub-text);
	}
	:global(::-webkit-scrollbar) { width: 4px; }
	:global(::-webkit-scrollbar-track) { background: transparent; }
	:global(::-webkit-scrollbar-thumb) { background: rgba(255,255,255,0.1); border-radius: 2px; }
</style>
