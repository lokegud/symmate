<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    variant?: 'elevated' | 'outlined' | 'flat' | 'glass';
    padding?: 'none' | 'sm' | 'md' | 'lg';
    children: Snippet;
    class?: string;
  }

  let {
    variant = 'elevated',
    padding = 'md',
    children,
    class: className = '',
  }: Props = $props();
</script>

<div class="card card--{variant} card--pad-{padding} {className}">
  {@render children()}
</div>

<style>
  .card {
    border-radius: var(--radius-lg);
    overflow: hidden;
  }

  .card--pad-none { padding: 0; }
  .card--pad-sm   { padding: var(--inset-sm); }
  .card--pad-md   { padding: var(--inset-md); }
  .card--pad-lg   { padding: var(--inset-lg); }

  .card--elevated {
    background: var(--color-surface);
    box-shadow: var(--shadow-md);
    border: 1px solid var(--color-border-subtle);
  }

  .card--outlined {
    background: var(--color-surface);
    border: 1.5px solid var(--color-border);
  }

  .card--flat {
    background: var(--color-surface-raised);
    border: none;
  }

  .card--glass {
    background: var(--color-surface-overlay);
    border: 1px solid var(--color-border-subtle);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: var(--shadow-lg);
  }
</style>
