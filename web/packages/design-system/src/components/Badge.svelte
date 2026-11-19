<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    variant?: 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'prism';
    size?: 'sm' | 'md';
    dot?: boolean;
    children: Snippet;
  }

  let {
    variant = 'neutral',
    size = 'md',
    dot = false,
    children,
  }: Props = $props();
</script>

<span class="badge badge--{variant} badge--{size}">
  {#if dot}
    <span class="badge__dot" aria-hidden="true"></span>
  {/if}
  {@render children()}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1-5);
    border-radius: var(--radius-sm);
    font-family: var(--font-body);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-wide);
    white-space: nowrap;
  }

  .badge--sm {
    height: 20px;
    padding: 0 var(--space-1-5);
    font-size: var(--text-xs);
  }
  .badge--md {
    height: 24px;
    padding: 0 var(--space-2);
    font-size: var(--text-sm);
  }

  .badge__dot {
    width: 6px;
    height: 6px;
    border-radius: var(--radius-full);
    background: currentColor;
    flex-shrink: 0;
  }

  .badge--success  { background: var(--color-success-subtle);  color: var(--color-success); }
  .badge--warning  { background: var(--color-warning-subtle);  color: var(--color-warning); }
  .badge--error    { background: var(--color-error-subtle);    color: var(--color-error); }
  .badge--info     { background: var(--color-info-subtle);     color: var(--color-info); }
  .badge--neutral  { background: var(--color-surface-raised);  color: var(--color-text-muted); }
  .badge--prism {
    background: var(--gradient-prism-subtle);
    color: var(--color-accent);
    border: 1px solid color-mix(in srgb, var(--color-accent) 20%, transparent);
  }
</style>
