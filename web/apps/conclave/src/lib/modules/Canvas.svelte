<script lang="ts">
	interface Note {
		id: string;
		x: number;
		y: number;
		text: string;
		done: boolean;
		color: string;
	}

	const NOTE_COLORS = ['#10b981', '#3d8ef0', '#8b5cf6', '#f59e0b', '#ef4444'];

	function load(): Note[] {
		try { return JSON.parse(localStorage.getItem('conclave:canvas') ?? '[]'); } catch { return []; }
	}
	function save(notes: Note[]) {
		try { localStorage.setItem('conclave:canvas', JSON.stringify(notes)); } catch {}
	}

	let notes = $state<Note[]>(load());
	let draggingNote = $state<string | null>(null);
	let dragOffset = $state({ x: 0, y: 0 });
	let editingId = $state<string | null>(null);

	function addNote(e: MouseEvent) {
		if ((e.target as HTMLElement).closest('.note')) return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const note: Note = {
			id: crypto.randomUUID(),
			x: e.clientX - rect.left,
			y: e.clientY - rect.top,
			text: '',
			done: false,
			color: NOTE_COLORS[notes.length % NOTE_COLORS.length]
		};
		notes = [...notes, note];
		save(notes);
		editingId = note.id;
	}

	function startDrag(e: MouseEvent, note: Note) {
		if (editingId === note.id) return;
		draggingNote = note.id;
		dragOffset = { x: e.clientX - note.x, y: e.clientY - note.y };
		e.preventDefault();
	}

	function onMouseMove(e: MouseEvent) {
		if (!draggingNote) return;
		notes = notes.map(n => n.id === draggingNote
			? { ...n, x: e.clientX - dragOffset.x, y: e.clientY - dragOffset.y }
			: n
		);
	}

	function onMouseUp() {
		if (draggingNote) { save(notes); draggingNote = null; }
	}

	function updateText(id: string, text: string) {
		notes = notes.map(n => n.id === id ? { ...n, text } : n);
		save(notes);
	}

	function toggleDone(id: string) {
		notes = notes.map(n => n.id === id ? { ...n, done: !n.done } : n);
		save(notes);
	}

	function removeNote(id: string) {
		notes = notes.filter(n => n.id !== id);
		save(notes);
	}
</script>

<svelte:window onmousemove={onMouseMove} onmouseup={onMouseUp} />

<div class="canvas" ondblclick={addNote} role="presentation">
	<div class="hint">double-click to add note</div>

	{#each notes as note (note.id)}
		<div
			class="note"
			class:done={note.done}
			style="left: {note.x}px; top: {note.y}px; --nc: {note.color};"
			onmousedown={(e) => startDrag(e, note)}
			role="note"
		>
			<div class="note-toolbar">
				<span class="note-dot" style="background: {note.color}"></span>
				<button class="note-btn" onclick={() => toggleDone(note.id)} title="toggle done">
					{note.done ? '↩' : '✓'}
				</button>
				<button class="note-btn" onclick={() => removeNote(note.id)} title="delete">×</button>
			</div>
			{#if editingId === note.id}
				<textarea
					value={note.text}
					oninput={(e) => updateText(note.id, (e.target as HTMLTextAreaElement).value)}
					onblur={() => editingId = null}
					placeholder="note…"
					autofocus
					onmousedown={(e) => e.stopPropagation()}
				></textarea>
			{:else}
				<div
					class="note-text"
					ondblclick={(e) => { e.stopPropagation(); editingId = note.id; }}
					role="presentation"
				>
					{note.text || '…'}
				</div>
			{/if}
		</div>
	{/each}
</div>

<style>
	.canvas {
		width: 100%;
		height: 100%;
		position: relative;
		overflow: hidden;
		cursor: crosshair;
	}

	.hint {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 11px;
		font-family: ui-monospace, monospace;
		color: rgba(255,255,255,0.1);
		pointer-events: none;
		letter-spacing: 0.05em;
	}

	.note {
		position: absolute;
		min-width: 140px;
		max-width: 240px;
		background: color-mix(in srgb, var(--nc) 10%, rgba(8,8,14,0.7));
		border: 1px solid color-mix(in srgb, var(--nc) 35%, transparent);
		border-radius: 6px;
		cursor: grab;
		user-select: none;
		transition: opacity 0.2s;
	}
	.note:active { cursor: grabbing; }
	.note.done { opacity: 0.45; }

	.note-toolbar {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 5px 7px;
		border-bottom: 1px solid rgba(255,255,255,0.06);
	}

	.note-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.note-btn {
		background: none;
		border: none;
		color: rgba(255,255,255,0.3);
		font-size: 12px;
		cursor: pointer;
		padding: 0 2px;
		line-height: 1;
		transition: color 0.15s;
		margin-left: auto;
	}
	.note-btn:first-of-type { margin-left: auto; }
	.note-btn:hover { color: rgba(255,255,255,0.7); }

	.note-text {
		padding: 8px 10px;
		font-size: 12px;
		font-family: ui-monospace, monospace;
		color: rgba(255,255,255,0.75);
		line-height: 1.5;
		white-space: pre-wrap;
		word-break: break-word;
		cursor: text;
		min-height: 32px;
	}

	textarea {
		width: 100%;
		min-height: 60px;
		background: transparent;
		border: none;
		outline: none;
		color: rgba(255,255,255,0.85);
		font-family: ui-monospace, monospace;
		font-size: 12px;
		padding: 8px 10px;
		resize: none;
		cursor: text;
	}
	textarea::placeholder { color: rgba(255,255,255,0.2); }
</style>
