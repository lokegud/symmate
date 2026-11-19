<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import '@symmate/design-system/minimal.css';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
