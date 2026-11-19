<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'warm';
    size?: 'sm' | 'md' | 'lg';
    disabled?: boolean;
    loading?: boolean;
    href?: string;
    type?: 'button' | 'submit' | 'reset';
    children: Snippet;
    onclick?: (e: MouseEvent) => void;
  }

  let {
    variant = 'primary',
    size = 'md',
    disabled = false,
    loading = false,
    href,
    type = 'button',
    children,
    onclick,
  }: Props = $props();

  const tag = href ? 'a' : 'button';
</script>

<svelte:element
  this={tag}
  {href}
  {type}
  disabled={disabled || loading}
  class="btn btn--{variant} btn--{size}"
  class:loading
  onclick={onclick}
  role={href ? 'link' : 'button'}
>
  {#if loading}
    <span class="btn__spinner" aria-hidden="true"></span>
  {/if}
  {@render children()}
</svelte:element>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    border: 1.5px solid transparent;
    border-radius: var(--radius-md);
    font-family: var(--font-body);
    font-weight: var(--weight-semibold);
    cursor: pointer;
    text-decoration: none;
    transition:
      background-color var(--duration-micro) var(--ease-out),
      color var(--duration-micro) var(--ease-out),
      border-color var(--duration-micro) var(--ease-out),
      box-shadow var(--duration-micro) var(--ease-out),
      transform var(--duration-micro) var(--ease-out);
    user-select: none;
    white-space: nowrap;
    position: relative;
  }

  .btn:focus-visible {
    outline: 2px solid var(--color-accent);
    outline-offset: 2px;
  }

  .btn:active:not(:disabled) {
    transform: translateY(1px);
  }

  .btn:disabled,
  .btn.loading {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
  }

  /* Sizes */
  .btn--sm {
    height: 32px;
    padding: 0 var(--inset-sm);
    font-size: var(--text-sm);
  }
  .btn--md {
    height: 40px;
    padding: 0 var(--inset-md);
    font-size: var(--text-base);
  }
  .btn--lg {
    height: 48px;
    padding: 0 var(--inset-lg);
    font-size: var(--text-md);
  }

  /* Variants */
  .btn--primary {
    background: var(--color-accent);
    color: var(--color-accent-text);
    border-color: var(--color-accent);
  }
  .btn--primary:hover:not(:disabled) {
    background: var(--color-accent-hover);
    border-color: var(--color-accent-hover);
  }
  .btn--primary:active:not(:disabled) {
    background: var(--color-accent-pressed);
    border-color: var(--color-accent-pressed);
  }

  .btn--secondary {
    background: var(--color-surface);
    color: var(--color-text);
    border-color: var(--color-border);
  }
  .btn--secondary:hover:not(:disabled) {
    background: var(--color-surface-raised);
    border-color: var(--color-border-strong);
  }

  .btn--ghost {
    background: transparent;
    color: var(--color-text);
    border-color: transparent;
  }
  .btn--ghost:hover:not(:disabled) {
    background: var(--color-surface-raised);
  }

  .btn--danger {
    background: var(--color-error);
    color: #fff;
    border-color: var(--color-error);
  }
  .btn--danger:hover:not(:disabled) {
    filter: brightness(1.1);
  }

  .btn--warm {
    background: var(--color-warm);
    color: var(--color-warm-text);
    border-color: var(--color-warm);
  }
  .btn--warm:hover:not(:disabled) {
    background: var(--color-warm-hover);
    border-color: var(--color-warm-hover);
  }

  /* Spinner */
  .btn__spinner {
    width: 14px;
    height: 14px;
    border: 2px solid currentColor;
    border-top-color: transparent;
    border-radius: var(--radius-full);
    animation: spin var(--duration-slow) linear infinite;
    flex-shrink: 0;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
