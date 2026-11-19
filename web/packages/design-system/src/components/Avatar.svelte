<script lang="ts">
  interface Props {
    src?: string;
    alt?: string;
    name?: string;
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
    ring?: boolean;
    ringColor?: 'accent' | 'success' | 'warning' | 'prism';
  }

  let {
    src,
    alt = '',
    name,
    size = 'md',
    ring = false,
    ringColor = 'accent',
  }: Props = $props();

  const initials = $derived(
    name
      ? name.split(' ').map(w => w[0]).slice(0, 2).join('').toUpperCase()
      : '?'
  );

  const sizes: Record<string, number> = {
    xs: 24, sm: 32, md: 40, lg: 56, xl: 72,
  };

  const px = $derived(sizes[size] ?? 40);
</script>

<span
  class="avatar avatar--{size}"
  class:avatar--ring={ring}
  class:avatar--ring-accent={ring && ringColor === 'accent'}
  class:avatar--ring-success={ring && ringColor === 'success'}
  class:avatar--ring-warning={ring && ringColor === 'warning'}
  class:avatar--ring-prism={ring && ringColor === 'prism'}
  style="--sz: {px}px"
  aria-label={name ?? alt}
  role="img"
>
  {#if src}
    <img src={src} alt={name ?? alt} class="avatar__img" />
  {:else}
    <span class="avatar__initials" aria-hidden="true">{initials}</span>
  {/if}
</span>

<style>
  .avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--sz);
    height: var(--sz);
    border-radius: var(--radius-full);
    background: var(--color-accent-subtle);
    overflow: hidden;
    flex-shrink: 0;
    position: relative;
  }

  .avatar__img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .avatar__initials {
    font-family: var(--font-body);
    font-weight: var(--weight-semibold);
    color: var(--color-accent);
    font-size: calc(var(--sz) * 0.36);
    line-height: 1;
  }

  /* Size overrides for font scale */
  .avatar--xs .avatar__initials { font-size: 9px; }
  .avatar--sm .avatar__initials { font-size: 12px; }

  /* Rings */
  .avatar--ring {
    padding: 2px;
  }
  .avatar--ring::before {
    content: '';
    position: absolute;
    inset: -2px;
    border-radius: inherit;
    padding: 2px;
    -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
  }
  .avatar--ring-accent::before  { background: var(--color-accent); }
  .avatar--ring-success::before { background: var(--color-success); }
  .avatar--ring-warning::before { background: var(--color-warning); }
  .avatar--ring-prism::before   { background: var(--gradient-prism); }
</style>
