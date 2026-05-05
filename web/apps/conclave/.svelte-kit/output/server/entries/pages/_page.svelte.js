import "../../chunks/index-server.js";
import { L as attr, R as escape_html, a as ensure_array_like, n as attr_class, r as attr_style, s as stringify } from "../../chunks/dev.js";
//#region src/lib/store.svelte.ts
var DEFAULTS = [
	{
		id: "hub",
		module: "hub",
		title: "Agents",
		tint: "#3d8ef0",
		x: 20,
		y: 20,
		w: 400,
		h: 500
	},
	{
		id: "canvas",
		module: "canvas",
		title: "Canvas",
		tint: "#10b981",
		x: 440,
		y: 20,
		w: 500,
		h: 500
	},
	{
		id: "chat",
		module: "chat",
		title: "Chat",
		tint: "#8b5cf6",
		x: 20,
		y: 540,
		w: 920,
		h: 340
	}
];
function loadBoxes() {
	try {
		const saved = localStorage.getItem("conclave:boxes");
		if (saved) return JSON.parse(saved);
	} catch {}
	return DEFAULTS;
}
function save(boxes) {
	try {
		localStorage.setItem("conclave:boxes", JSON.stringify(boxes));
	} catch {}
}
var BoxStore = class {
	boxes = loadBoxes();
	update(id, patch) {
		this.boxes = this.boxes.map((b) => b.id === id ? {
			...b,
			...patch
		} : b);
		save(this.boxes);
	}
	add(b) {
		this.boxes = [...this.boxes, b];
		save(this.boxes);
	}
	remove(id) {
		this.boxes = this.boxes.filter((b) => b.id !== id);
		save(this.boxes);
	}
	bringToFront(id) {
		const box = this.boxes.find((b) => b.id === id);
		if (!box) return;
		this.boxes = [...this.boxes.filter((b) => b.id !== id), box];
		save(this.boxes);
	}
};
var boxStore = new BoxStore();
var AgentStore = class {
	agents = [];
	set(agents) {
		this.agents = agents;
	}
};
var agentStore = new AgentStore();
//#endregion
//#region src/lib/Box.svelte
function Box($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let { box, zIndex, children } = $$props;
		$$renderer.push(`<div class="box svelte-q7938w"${attr_style(` left: ${stringify(box.x)}px; top: ${stringify(box.y)}px; width: ${stringify(box.w)}px; height: ${stringify(box.h)}px; --tint: ${stringify(box.tint)}; z-index: ${stringify(zIndex)}; `)} role="region"${attr("aria-label", box.title)}><div class="titlebar svelte-q7938w"><span class="dot svelte-q7938w"${attr_style(`background: ${stringify(box.tint)}`)}></span> <span class="title svelte-q7938w">${escape_html(box.title)}</span> <button class="close svelte-q7938w">×</button></div> <div class="content svelte-q7938w">`);
		children($$renderer);
		$$renderer.push(`<!----></div> <div class="resize-handle svelte-q7938w" role="presentation"></div></div>`);
	});
}
//#endregion
//#region src/lib/modules/AgentHub.svelte
function AgentHub($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		const STATUS_COLOR = {
			idle: "#6b7280",
			working: "#10b981",
			blocked: "#f59e0b",
			done: "#3d8ef0"
		};
		let selectedAgent = null;
		$$renderer.push(`<div class="hub svelte-gj8al9">`);
		if (agentStore.agents.length === 0) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<div class="empty svelte-gj8al9">No agents registered</div>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--[-->`);
			const each_array = ensure_array_like(agentStore.agents);
			for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
				let agent = each_array[$$index_1];
				$$renderer.push(`<div${attr_class("agent svelte-gj8al9", void 0, { "selected": selectedAgent === agent.id })} role="button" tabindex="0"><div class="agent-header svelte-gj8al9"><span class="status-dot svelte-gj8al9"${attr_style(`background: ${stringify(STATUS_COLOR[agent.status])}`)}></span> <span class="agent-name svelte-gj8al9">${escape_html(agent.name)}</span> <span class="status-label svelte-gj8al9">${escape_html(agent.status)}</span></div> <div class="agent-task svelte-gj8al9">${escape_html(agent.task || "No task")}</div> `);
				if (selectedAgent === agent.id && agent.todos?.length > 0) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="todos svelte-gj8al9"><!--[-->`);
					const each_array_1 = ensure_array_like(agent.todos);
					for (let $$index = 0, $$length = each_array_1.length; $$index < $$length; $$index++) {
						let todo = each_array_1[$$index];
						$$renderer.push(`<div${attr_class("todo svelte-gj8al9", void 0, { "done": todo.done })}><span class="todo-check svelte-gj8al9">${escape_html(todo.done ? "✓" : "○")}</span> <span>${escape_html(todo.text)}</span></div>`);
					}
					$$renderer.push(`<!--]--></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></div>`);
			}
			$$renderer.push(`<!--]-->`);
		}
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
//#region src/lib/modules/Chat.svelte
function Chat($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let messages = [];
		let input = "";
		let selectedAgentId = "";
		function agentName(id) {
			return agentStore.agents.find((a) => a.id === id)?.name ?? id;
		}
		$$renderer.push(`<div class="chat svelte-4s3qff"><div class="agent-select svelte-4s3qff">`);
		if (agentStore.agents.length === 0) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<span class="no-agents svelte-4s3qff">No agents online</span>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--[-->`);
			const each_array = ensure_array_like(agentStore.agents);
			for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
				let agent = each_array[$$index];
				$$renderer.push(`<button${attr_class("agent-chip svelte-4s3qff", void 0, { "active": selectedAgentId === agent.id })}>${escape_html(agent.name)}</button>`);
			}
			$$renderer.push(`<!--]-->`);
		}
		$$renderer.push(`<!--]--></div> <div class="messages svelte-4s3qff"><!--[-->`);
		const each_array_1 = ensure_array_like(messages);
		for (let i = 0, $$length = each_array_1.length; i < $$length; i++) {
			let msg = each_array_1[i];
			$$renderer.push(`<div${attr_class(`msg ${stringify(msg.role)}`, "svelte-4s3qff")}>`);
			if (msg.role === "assistant") {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<span class="msg-from svelte-4s3qff">${escape_html(agentName(msg.agentId ?? ""))}</span>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--> <span class="msg-text svelte-4s3qff">${escape_html(msg.content)}${escape_html((msg.role, ""))}</span></div>`);
		}
		$$renderer.push(`<!--]--> <div></div></div> <div class="input-row svelte-4s3qff"><textarea${attr("placeholder", "Select an agent")}${attr("disabled", true, true)} rows="2" class="svelte-4s3qff">`);
		const $$body = escape_html(input);
		if ($$body) $$renderer.push(`${$$body}`);
		$$renderer.push(`</textarea> <button${attr("disabled", !input.trim(), true)} class="svelte-4s3qff">↑</button></div></div>`);
	});
}
//#endregion
//#region src/lib/modules/Canvas.svelte
function Canvas($$renderer) {
	function load() {
		try {
			return JSON.parse(localStorage.getItem("conclave:canvas") ?? "[]");
		} catch {
			return [];
		}
	}
	let notes = load();
	let editingId = null;
	$$renderer.push(`<div class="canvas svelte-1opbyn1" role="presentation"><div class="hint svelte-1opbyn1">double-click to add note</div> <!--[-->`);
	const each_array = ensure_array_like(notes);
	for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
		let note = each_array[$$index];
		$$renderer.push(`<div${attr_class("note svelte-1opbyn1", void 0, { "done": note.done })}${attr_style(`left: ${stringify(note.x)}px; top: ${stringify(note.y)}px; --nc: ${stringify(note.color)};`)} role="note"><div class="note-toolbar svelte-1opbyn1"><span class="note-dot svelte-1opbyn1"${attr_style(`background: ${stringify(note.color)}`)}></span> <button class="note-btn svelte-1opbyn1" title="toggle done">${escape_html(note.done ? "↩" : "✓")}</button> <button class="note-btn svelte-1opbyn1" title="delete">×</button></div> `);
		if (editingId === note.id) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<textarea placeholder="note…" autofocus="" class="svelte-1opbyn1">`);
			const $$body = escape_html(note.text);
			if ($$body) $$renderer.push(`${$$body}`);
			$$renderer.push(`</textarea>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<div class="note-text svelte-1opbyn1" role="presentation">${escape_html(note.text || "…")}</div>`);
		}
		$$renderer.push(`<!--]--></div>`);
	}
	$$renderer.push(`<!--]--></div>`);
}
//#endregion
//#region src/routes/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		$$renderer.push(`<div class="shell svelte-1uha8ag"><div class="grid-bg svelte-1uha8ag"></div> <!--[-->`);
		const each_array = ensure_array_like(boxStore.boxes);
		for (let i = 0, $$length = each_array.length; i < $$length; i++) {
			let box = each_array[i];
			Box($$renderer, {
				box,
				zIndex: 100 + i,
				children: ($$renderer) => {
					if (box.module === "hub") {
						$$renderer.push("<!--[0-->");
						AgentHub($$renderer, {});
					} else if (box.module === "chat") {
						$$renderer.push("<!--[1-->");
						Chat($$renderer, {});
					} else if (box.module === "canvas") {
						$$renderer.push("<!--[2-->");
						Canvas($$renderer, {});
					} else $$renderer.push("<!--[-1-->");
					$$renderer.push(`<!--]-->`);
				},
				$$slots: { default: true }
			});
		}
		$$renderer.push(`<!--]--> <div class="toolbar svelte-1uha8ag"><span class="wordmark svelte-1uha8ag">⬡ conclave</span> <div class="toolbar-actions svelte-1uha8ag">`);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<button class="btn-new svelte-1uha8ag">+ Box</button>`);
		$$renderer.push(`<!--]--></div> <div class="hub-status svelte-1uha8ag"><span${attr_class("dot svelte-1uha8ag", void 0, { "connected": agentStore.agents.length >= 0 })}></span> <span>${escape_html(agentStore.agents.length)} agent${escape_html(agentStore.agents.length !== 1 ? "s" : "")}</span></div></div></div>`);
	});
}
//#endregion
export { _page as default };
