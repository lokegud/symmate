<script lang="ts">
	import type { BoxState } from './store.svelte.js';
	import { boxStore } from './store.svelte.js';

	let { box, zIndex, children }: {
		box: BoxState;
		zIndex: number;
		children: import('svelte').Snippet;
	} = $props();

	let dragging = false;
	let resizing = false;
	let dragStart = { mx: 0, my: 0, bx: 0, by: 0 };
	let resizeStart = { mx: 0, my: 0, bw: 0, bh: 0 };

	function onDragStart(e: MouseEvent) {
		if ((e.target as HTMLElement).closest('.resize-handle')) return;
		dragging = true;
		dragStart = { mx: e.clientX, my: e.clientY, bx: box.x, by: box.y };
		boxStore.bringToFront(box.id);
		e.preventDefault();
	}

	function onResizeStart(e: MouseEvent) {
		resizing = true;
		resizeStart = { mx: e.clientX, my: e.clientY, bw: box.w, bh: box.h };
		e.stopPropagation();
		e.preventDefault();
	}

	function onMouseMove(e: MouseEvent) {
		if (dragging) {
			boxStore.update(box.id, {
				x: Math.max(0, dragStart.bx + e.clientX - dragStart.mx),
				y: Math.max(0, dragStart.by + e.clientY - dragStart.my)
			});
		}
		if (resizing) {
			boxStore.update(box.id, {
				w: Math.max(200, resizeStart.bw + e.clientX - resizeStart.mx),
				h: Math.max(150, resizeStart.bh + e.clientY - resizeStart.my)
			});
		}
	}

	function onMouseUp() {
		dragging = false;
		resizing = false;
	}
</script>

<svelte:window onmousemove={onMouseMove} onmouseup={onMouseUp} />

<div
	class="box"
	style="
		left: {box.x}px; top: {box.y}px;
		width: {box.w}px; height: {box.h}px;
		--tint: {box.tint};
		z-index: {zIndex};
	"
	onmousedown={onDragStart}
	role="region"
	aria-label={box.title}
>
	<div class="titlebar">
		<span class="dot" style="background: {box.tint}"></span>
		<span class="title">{box.title}</span>
		<button class="close" onclick={() => boxStore.remove(box.id)}>×</button>
	</div>
	<div class="content">
		{@render children()}
	</div>
	<div class="resize-handle" onmousedown={onResizeStart} role="presentation"></div>
</div>

<style>
	.box {
		position: absolute;
		background: color-mix(in srgb, var(--tint) 6%, var(--hub-glass-bg));
		backdrop-filter: blur(var(--hub-glass-blur));
		-webkit-backdrop-filter: blur(var(--hub-glass-blur));
		border: 1px solid color-mix(in srgb, var(--tint) 30%, var(--hub-border));
		border-radius: var(--hub-radius-lg);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-shadow: var(--hub-shadow-box);
		user-select: none;
	}

	.titlebar {
		display: flex;
		align-items: center;
		gap: var(--hub-space-2);
		padding: var(--hub-space-2) var(--hub-space-3);
		cursor: grab;
		border-bottom: 1px solid color-mix(in srgb, var(--tint) 15%, var(--hub-border-subtle));
		background: color-mix(in srgb, var(--tint) 8%, rgba(0,0,0,0.2));
		flex-shrink: 0;
	}

	.titlebar:active { cursor: grabbing; }

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		opacity: 0.8;
		flex-shrink: 0;
	}

	.title {
		font-family: var(--hub-font-mono);
		font-size: var(--hub-text-xs);
		font-weight: 500;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: color-mix(in srgb, var(--tint) 70%, rgba(255,255,255,0.7));
		flex: 1;
	}

	.close {
		background: none;
		border: none;
		color: var(--hub-text-subtle);
		font-size: 16px;
		line-height: 1;
		cursor: pointer;
		padding: 0 2px;
		transition: color var(--hub-duration-fast);
	}
	.close:hover { color: var(--hub-text); }

	.content {
		flex: 1;
		overflow: auto;
		min-height: 0;
	}

	.resize-handle {
		position: absolute;
		bottom: 0;
		right: 0;
		width: 16px;
		height: 16px;
		cursor: se-resize;
	}
	.resize-handle::after {
		content: '';
		position: absolute;
		bottom: 3px;
		right: 3px;
		width: 8px;
		height: 8px;
		border-right: 2px solid rgba(255,255,255,0.15);
		border-bottom: 2px solid rgba(255,255,255,0.15);
	}
</style>
