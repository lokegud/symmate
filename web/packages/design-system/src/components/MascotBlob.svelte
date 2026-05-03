<script lang="ts">
  import type { BlobState, BlobMode, BlobSize } from '../tokens.js';

  interface Props {
    seed?: string;
    state?: BlobState;
    size?: BlobSize;
    evolved?: boolean;
    mode?: BlobMode;
    animated?: boolean;
    label?: string;
  }

  let {
    seed = 'platform-canonical',
    state = 'idle',
    size = 'md',
    evolved = false,
    mode = 'per',
    animated = true,
    label,
  }: Props = $props();

  /* Derive visual properties from seed — deterministic, no randomness at render */
  function hashSeed(s: string): number {
    let h = 0;
    for (let i = 0; i < s.length; i++) {
      h = ((h << 5) - h + s.charCodeAt(i)) | 0;
    }
    return Math.abs(h);
  }

  const seedHash = $derived(hashSeed(seed));

  /* Mark density: 0–1. Low = new instance, high = seasoned */
  const markDensity = $derived(
    evolved
      ? 0.6 + (seedHash % 40) / 100   /* 0.60–1.00 */
      : 0.1 + (seedHash % 30) / 100   /* 0.10–0.40 */
  );

  /* Color temperature: 0 = cool (analytical), 1 = warm (optimistic) */
  const colorTemp = $derived((seedHash % 100) / 100);

  /* Per-seed gradient angle variation */
  const gradAngle = $derived(125 + (seedHash % 40));

  /* Mark count (number of internal circles) */
  const markCount = $derived(Math.floor(markDensity * 8) + 2);

  /* Sizes in px */
  const sizePx: Record<BlobSize, number> = { sm: 64, md: 96, lg: 128, xl: 192 };
  const px = $derived(sizePx[size] ?? 96);

  /* Border-radius shapes per state — the Blob takes different resting shapes */
  const stateShapes: Record<BlobState, string> = {
    idle:        '60% 40% 55% 45% / 50% 45% 55% 50%',
    thinking:    '50% 50% 40% 60% / 60% 40% 55% 45%',
    listening:   '55% 45% 60% 40% / 45% 60% 45% 55%',
    happy:       '65% 35% 60% 40% / 55% 50% 50% 55%',
    content:     '58% 42% 52% 48% / 48% 52% 48% 52%',
    nervous:     '45% 55% 50% 50% / 55% 45% 55% 45%',
    reluctant:   '52% 48% 48% 52% / 42% 58% 48% 52%',
    guarding:    '48% 52% 55% 45% / 52% 48% 52% 48%',
    disagreeing: '50% 50% 50% 50% / 50% 50% 50% 50%',
    dormant:     '65% 35% 65% 35% / 35% 65% 35% 65%',
  };

  const shape = $derived(stateShapes[state]);

  /* Eye expression per state */
  const eyeStates: Record<BlobState, { scaleY: number; offsetY: number; eyeClose?: boolean }> = {
    idle:        { scaleY: 1.0,  offsetY: 0 },
    thinking:    { scaleY: 0.8,  offsetY: -2 },
    listening:   { scaleY: 1.1,  offsetY: 0 },
    happy:       { scaleY: 0.6,  offsetY: 2 },    /* squinting happy */
    content:     { scaleY: 0.65, offsetY: 1 },    /* half-lidded */
    nervous:     { scaleY: 1.3,  offsetY: -1 },   /* wide */
    reluctant:   { scaleY: 0.9,  offsetY: -1 },   /* slightly angled down */
    guarding:    { scaleY: 1.15, offsetY: -2 },   /* alert wide */
    disagreeing: { scaleY: 1.0,  offsetY: 0 },    /* direct eye contact — same as idle but posture shifts */
    dormant:     { scaleY: 0.15, offsetY: 0, eyeClose: true },
  };

  const eyeState = $derived(eyeStates[state]);

  /* Mark animation speed — faster when active */
  const activeStates = new Set<BlobState>(['thinking', 'guarding', 'disagreeing', 'nervous']);
  const markSpeed = $derived(activeStates.has(state) ? '1.5s' : '3.5s');

  /* Warm gradient shift per color temperature and state */
  const warmFactor = $derived(
    state === 'happy' || state === 'content' ? Math.min(colorTemp + 0.3, 1) : colorTemp
  );

  /* Dormant: nearly transparent */
  const opacity = $derived(state === 'dormant' ? 0.4 : 1);

  /* Aria label */
  const ariaLabel = $derived(label ?? `Symmate blob: ${state}`);

  /* Generate mark positions deterministically from seed + count */
  function genMarks(count: number, hash: number) {
    return Array.from({ length: count }, (_, i) => ({
      cx: 20 + ((hash * (i + 1) * 37) % 60),
      cy: 20 + ((hash * (i + 1) * 53) % 60),
      r:  2 + ((hash * (i + 7)) % 5),
      delay: ((hash * i * 13) % 20) / 10,
    }));
  }

  const marks = $derived(genMarks(markCount, seedHash));

  /* Gradient color interpolation for warm/cool axis */
  function lerpHex(a: [number,number,number], b: [number,number,number], t: number): string {
    const r = Math.round(a[0] + (b[0]-a[0])*t);
    const g = Math.round(a[1] + (b[1]-a[1])*t);
    const bl = Math.round(a[2] + (b[2]-a[2])*t);
    return `rgb(${r},${g},${bl})`;
  }

  /* Crown: blue → violet based on warm factor */
  const crownColor = $derived(lerpHex([75,95,211], [139,91,211], warmFactor * 0.5));
  /* Mid: violet */
  const midColor   = '#8B5BD3';
  /* Base: amber → warm amber */
  const baseColor  = $derived(lerpHex([212,137,90], [232,168,120], warmFactor));
</script>

<figure
  class="blob blob--{size} blob--{state}"
  class:blob--animated={animated}
  class:blob--evolved={evolved}
  style="
    --blob-shape: {shape};
    --blob-px: {px}px;
    --blob-angle: {gradAngle}deg;
    --blob-crown: {crownColor};
    --blob-mid: {midColor};
    --blob-base: {baseColor};
    --blob-mark-speed: {markSpeed};
    --blob-opacity: {opacity};
    --blob-eye-scale: {eyeState.scaleY};
    --blob-eye-offset: {eyeState.offsetY}px;
  "
  role="img"
  aria-label={ariaLabel}
>
  <!-- Blob body -->
  <div class="blob__body">
    <!-- Internal marks (SVG, clipped to body shape) -->
    <svg class="blob__marks" viewBox="0 0 100 100" aria-hidden="true">
      {#each marks as mark, i}
        <circle
          cx={mark.cx}
          cy={mark.cy}
          r={mark.r}
          class="blob__mark"
          style="animation-delay: -{mark.delay}s"
          opacity={0.35 + markDensity * 0.4}
        />
      {/each}
    </svg>

    <!-- Eyes -->
    <div class="blob__face" aria-hidden="true">
      <div class="blob__eye blob__eye--left">
        <div class="blob__pupil"></div>
        <div class="blob__sparkle"></div>
      </div>
      <div class="blob__eye blob__eye--right">
        <div class="blob__pupil"></div>
        <div class="blob__sparkle"></div>
      </div>
    </div>

    <!-- Thinking indicator -->
    {#if state === 'thinking'}
      <div class="blob__thinking" aria-hidden="true">
        <span></span><span></span><span></span>
      </div>
    {/if}
  </div>
</figure>

<style>
  .blob {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    margin: 0;
    flex-shrink: 0;
    position: relative;
    opacity: var(--blob-opacity);
    transition: opacity var(--duration-slow) var(--ease-inout);
  }

  .blob--sm  { --blob-eye-size: 10px; }
  .blob--md  { --blob-eye-size: 14px; }
  .blob--lg  { --blob-eye-size: 18px; }
  .blob--xl  { --blob-eye-size: 26px; }

  .blob__body {
    width: var(--blob-px);
    height: var(--blob-px);
    border-radius: var(--blob-shape);
    background: linear-gradient(var(--blob-angle),
      var(--blob-crown) 0%,
      var(--blob-mid)   50%,
      var(--blob-base)  100%);
    position: relative;
    overflow: hidden;
    box-shadow: var(--shadow-prism);
    transition:
      border-radius var(--duration-blob) var(--ease-spring),
      background var(--duration-slow) var(--ease-inout),
      box-shadow var(--duration-normal) var(--ease-out);
  }

  /* Idle breathing animation */
  .blob--animated .blob__body {
    animation: blob-breathe var(--duration-pulse) var(--ease-inout) infinite;
  }

  /* State overrides pause the breathe and apply state shape */
  .blob--thinking .blob__body,
  .blob--guarding .blob__body,
  .blob--disagreeing .blob__body {
    animation-play-state: paused;
  }

  /* Marks */
  .blob__marks {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }

  .blob__mark {
    fill: white;
    animation: mark-drift var(--blob-mark-speed) var(--ease-inout) infinite alternate;
  }

  @keyframes mark-drift {
    0%   { transform: translate(0, 0); }
    100% { transform: translate(4px, -3px); }
  }

  /* Face */
  .blob__face {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: calc(var(--blob-px) * 0.14);
    padding-top: calc(var(--blob-px) * 0.05);
  }

  .blob__eye {
    width: var(--blob-eye-size);
    height: var(--blob-eye-size);
    background: #fff;
    border-radius: var(--radius-full);
    position: relative;
    transform: scaleY(var(--blob-eye-scale)) translateY(var(--blob-eye-offset));
    transition: transform var(--duration-moderate) var(--ease-spring);
  }

  .blob__pupil {
    width: 55%;
    height: 55%;
    background: #1C1A16;
    border-radius: var(--radius-full);
    position: absolute;
    bottom: 20%;
    left: 50%;
    transform: translateX(-50%);
  }

  .blob__sparkle {
    width: 25%;
    height: 25%;
    background: white;
    border-radius: var(--radius-full);
    position: absolute;
    top: 15%;
    right: 15%;
  }

  /* Thinking dots */
  .blob__thinking {
    position: absolute;
    bottom: 18%;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    gap: 3px;
    align-items: center;
  }

  .blob__thinking span {
    width: 4px;
    height: 4px;
    border-radius: var(--radius-full);
    background: rgba(255,255,255,0.7);
    animation: thinking-pulse var(--duration-slow) ease-in-out infinite;
  }

  .blob__thinking span:nth-child(2) { animation-delay: 0.15s; }
  .blob__thinking span:nth-child(3) { animation-delay: 0.30s; }

  @keyframes thinking-pulse {
    0%, 80%, 100% { transform: scale(0.8); opacity: 0.5; }
    40%           { transform: scale(1.2); opacity: 1; }
  }

  /* Guarding — subtle glow intensification */
  .blob--guarding .blob__body {
    box-shadow: var(--shadow-prism), 0 0 12px rgba(75,95,211,0.4);
  }

  /* Dormant — marks slow down, applied via --blob-mark-speed above */
  .blob--dormant .blob__body {
    filter: saturate(0.6) brightness(0.8);
  }

  /* Evolved indicator ring */
  .blob--evolved .blob__body::after {
    content: '';
    position: absolute;
    inset: -2px;
    border-radius: inherit;
    background: var(--gradient-prism);
    z-index: -1;
    opacity: 0.6;
    animation: prism-shift 4s ease-in-out infinite alternate;
  }

  @keyframes prism-shift {
    0%   { background-position: 0% 50%; }
    100% { background-position: 100% 50%; }
  }

  @media (prefers-reduced-motion: reduce) {
    .blob--animated .blob__body {
      animation: none;
    }
    .blob__mark {
      animation: none;
    }
  }
</style>
