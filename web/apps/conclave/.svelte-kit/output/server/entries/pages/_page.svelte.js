import { B as attr, V as escape_html, a as attr_style, i as attr_class, l as stringify, o as derived, s as ensure_array_like, u as html } from "../../chunks/index-server.js";
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
typeof window !== "undefined" && window.__TAURI__;
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
		let members = [];
		let selectedId = null;
		function statusColor(status) {
			if (status === "working") return "var(--state-working)";
			if (status === "offline") return "var(--state-revoked)";
			return "var(--state-idle)";
		}
		function relativeTime(iso) {
			const diff = Date.now() - new Date(iso).getTime();
			const mins = Math.floor(diff / 6e4);
			if (mins < 1) return "just now";
			if (mins < 60) return `${mins}m ago`;
			return `${Math.floor(mins / 60)}h ago`;
		}
		const display = derived(() => members.length > 0 ? members : agentStore.agents.map((a) => ({
			id: a.id,
			name: a.name,
			fingerprint: "",
			bindrune_url: "",
			status: a.status === "working" ? "working" : "idle",
			task: a.task,
			pane_cmd: "",
			last_seen: a.last_seen,
			source: "sygil"
		})));
		$$renderer.push(`<div class="hub svelte-gj8al9">`);
		if (display().length === 0) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<div class="empty svelte-gj8al9">${escape_html("No agents registered")}</div>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--[-->`);
			const each_array = ensure_array_like(display());
			for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
				let m = each_array[$$index];
				$$renderer.push(`<div${attr_class("member svelte-gj8al9", void 0, { "selected": selectedId === m.id })} role="button" tabindex="0"><div class="member-header svelte-gj8al9">`);
				if (m.bindrune_url) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<img class="bindrune svelte-gj8al9"${attr("src", m.bindrune_url)}${attr("alt", m.name)}/>`);
				} else {
					$$renderer.push("<!--[-1-->");
					$$renderer.push(`<span class="bindrune-placeholder svelte-gj8al9">${escape_html(m.name[0])}</span>`);
				}
				$$renderer.push(`<!--]--> <div class="member-meta svelte-gj8al9"><div class="name-row svelte-gj8al9"><span class="member-name svelte-gj8al9">${escape_html(m.name)}</span> <span class="source-badge svelte-gj8al9"${attr("data-source", m.source)}>${escape_html(m.source)}</span></div> <div class="task-row svelte-gj8al9"><span class="status-dot svelte-gj8al9"${attr_style(`background: ${stringify(statusColor(m.status))}`)}${attr("title", m.status)}></span> <span class="task-text svelte-gj8al9">${escape_html(m.task || m.status)}</span></div></div> <span class="last-seen svelte-gj8al9">${escape_html(relativeTime(m.last_seen))}</span></div> `);
				if (selectedId === m.id && m.pane_cmd) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="pane-cmd svelte-gj8al9"><span class="pane-label svelte-gj8al9">pane</span> <code class="svelte-gj8al9">${escape_html(m.pane_cmd)}</code></div>`);
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
//#region src/lib/modules/CohortPresence.svelte
function CohortPresence($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		const STATE_COLOR = {
			idle: "var(--state-idle)",
			working: "var(--state-working)",
			blocked: "var(--state-blocked)",
			done: "var(--state-done)"
		};
		const STATE_AURA = {
			working: "var(--state-working-aura)",
			blocked: "var(--state-blocked-aura)",
			done: "var(--state-done-aura)",
			idle: "transparent"
		};
		let selected = null;
		let bindruneSvgs = {};
		function initials(name) {
			return name.slice(0, 2).toUpperCase();
		}
		function formatTime(iso) {
			try {
				return new Date(iso).toLocaleTimeString([], {
					hour: "2-digit",
					minute: "2-digit"
				});
			} catch {
				return "—";
			}
		}
		$$renderer.push(`<div class="presence svelte-1980tut">`);
		if (agentStore.agents.length === 0) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<div class="empty svelte-1980tut">Awaiting cohort registration…</div>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--[-->`);
			const each_array = ensure_array_like(agentStore.agents);
			for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
				let agent = each_array[$$index_1];
				const color = STATE_COLOR[agent.status] ?? STATE_COLOR.idle;
				const aura = STATE_AURA[agent.status] ?? "transparent";
				const open = selected === agent.id;
				$$renderer.push(`<button${attr_class("card svelte-1980tut", void 0, { "open": open })}${attr_style(`--state-color: ${stringify(color)}; --state-aura: ${stringify(aura)};`)}${attr("aria-label", `${stringify(agent.name)} — ${stringify(agent.status)}`)}><div${attr_class("mark svelte-1980tut", void 0, { "pulsing": agent.status === "working" })}>`);
				if (bindruneSvgs[agent.id]) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="bindrune-svg svelte-1980tut">${html(bindruneSvgs[agent.id])}</div>`);
				} else {
					$$renderer.push("<!--[-1-->");
					$$renderer.push(`<div class="monogram svelte-1980tut">${escape_html(initials(agent.name))}</div>`);
				}
				$$renderer.push(`<!--]--> <div class="mark-ring svelte-1980tut"></div></div> <div class="info svelte-1980tut"><div class="name-row svelte-1980tut"><span class="name svelte-1980tut">${escape_html(agent.name)}</span> <span class="status-chip svelte-1980tut">${escape_html(agent.status)}</span></div> <div class="task svelte-1980tut">${escape_html(agent.task || "No active task")}</div> `);
				if (open) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="expanded svelte-1980tut"><div class="last-seen svelte-1980tut">Last seen ${escape_html(formatTime(agent.last_seen))}</div> `);
					if (agent.todos?.length > 0) {
						$$renderer.push("<!--[0-->");
						$$renderer.push(`<ul class="todos svelte-1980tut"><!--[-->`);
						const each_array_1 = ensure_array_like(agent.todos);
						for (let $$index = 0, $$length = each_array_1.length; $$index < $$length; $$index++) {
							let todo = each_array_1[$$index];
							$$renderer.push(`<li${attr_class("todo svelte-1980tut", void 0, { "done": todo.done })}><span class="todo-mark svelte-1980tut">${escape_html(todo.done ? "✓" : "·")}</span> ${escape_html(todo.text)}</li>`);
						}
						$$renderer.push(`<!--]--></ul>`);
					} else $$renderer.push("<!--[-1-->");
					$$renderer.push(`<!--]--></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></div></button>`);
			}
			$$renderer.push(`<!--]-->`);
		}
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
//#region src/lib/modules/ConversationStream.svelte
function ConversationStream($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let messages = [];
		function shortSender(sender) {
			return sender.replace(/^@/, "").split(":")[0];
		}
		function formatTime(iso) {
			try {
				return new Date(iso).toLocaleTimeString([], {
					hour: "2-digit",
					minute: "2-digit"
				});
			} catch {
				return "??:??";
			}
		}
		function isPause(body) {
			return /^\s*PAUSE\b/i.test(body);
		}
		function isRoundRobin(body, sender) {
			return /^@?elder/.test(sender) && /\*\*.*cohort|round|task \d|change [A-Z]/i.test(body);
		}
		function isLoke(sender) {
			return /^@?loke/.test(sender);
		}
		$$renderer.push(`<div class="stream svelte-z7v866">`);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <div class="messages svelte-z7v866">`);
		if (messages.length === 0 && true) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<!--[-->`);
			const each_array = ensure_array_like(Array(6));
			for (let i = 0, $$length = each_array.length; i < $$length; i++) {
				each_array[i];
				$$renderer.push(`<div class="skeleton-row svelte-z7v866"${attr_style(`opacity: ${stringify(.3 + i % 2 * .1)}`)}></div>`);
			}
			$$renderer.push(`<!--]-->`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--[-->`);
			const each_array_1 = ensure_array_like(messages);
			for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
				let msg = each_array_1[$$index_1];
				const sender = shortSender(msg.sender);
				const pause = isPause(msg.body);
				const rr = !pause && isRoundRobin(msg.body, sender);
				const loke = !pause && !rr && isLoke(msg.sender);
				$$renderer.push(`<div${attr_class("row svelte-z7v866", void 0, {
					"is-pause": pause,
					"is-rr": rr,
					"is-loke": loke
				})}><span class="ts svelte-z7v866">${escape_html(formatTime(msg.ts))}</span> <span class="mark svelte-z7v866"${attr("title", msg.sender)}>${escape_html(sender.slice(0, 2).toUpperCase())}</span> <span class="sender svelte-z7v866">${escape_html(sender)}</span> <span class="body svelte-z7v866">${escape_html(msg.body)}</span></div>`);
			}
			$$renderer.push(`<!--]-->`);
		}
		$$renderer.push(`<!--]--> <div></div></div> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
//#region src/lib/modules/DecisionLog.svelte
function DecisionLog($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		const AUTHOR_COLOR = {
			elder: "#e8a878",
			keel: "#4a9eff",
			quill: "#8b5cf6",
			trace: "#f59e0b",
			wren: "#10b981",
			wright: "#3d8ef0",
			forge: "#ec4899",
			loke: "rgba(255,255,255,0.7)"
		};
		function authorColor(author) {
			return AUTHOR_COLOR[author.toLowerCase()] ?? "var(--hub-text-muted)";
		}
		let decisions = [];
		let expanded = /* @__PURE__ */ new Set();
		function formatDate(iso) {
			try {
				return new Date(iso).toLocaleDateString([], {
					month: "short",
					day: "numeric",
					year: "numeric"
				});
			} catch {
				return iso;
			}
		}
		$$renderer.push(`<div class="log svelte-1spkd1">`);
		if (decisions.length === 0) {
			$$renderer.push("<!--[1-->");
			$$renderer.push(`<div class="empty-state svelte-1spkd1">No decisions logged yet.<br/> Append to <code class="svelte-1spkd1">/mnt/KB/claude/_decisions/symmate-build.md</code>.</div>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<div class="timeline svelte-1spkd1"><!--[-->`);
			const each_array = ensure_array_like(decisions);
			for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
				let d = each_array[$$index];
				const open = expanded.has(d.line_number);
				const color = authorColor(d.author);
				$$renderer.push(`<div class="entry svelte-1spkd1"><div class="rail svelte-1spkd1"><div class="rail-dot svelte-1spkd1"${attr_style(`background: ${stringify(color)}; box-shadow: 0 0 6px ${stringify(color)}33`)}></div> <div class="rail-line svelte-1spkd1"></div></div> <div class="content svelte-1spkd1"><div class="meta svelte-1spkd1"><span class="date svelte-1spkd1">${escape_html(formatDate(d.date))}</span> <span class="author svelte-1spkd1"${attr_style(`color: ${stringify(color)}`)}>${escape_html(d.author)}</span></div> <div class="title svelte-1spkd1">${escape_html(d.title)}</div> `);
				if (d.body && d.body.trim()) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<button class="expand-btn svelte-1spkd1"${attr("aria-expanded", open)}>${escape_html(open ? "▲ collapse" : "▼ expand")}</button> `);
					if (open) {
						$$renderer.push("<!--[0-->");
						$$renderer.push(`<pre class="body svelte-1spkd1">${escape_html(d.body.trim())}</pre>`);
					} else $$renderer.push("<!--[-1-->");
					$$renderer.push(`<!--]-->`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></div></div>`);
			}
			$$renderer.push(`<!--]--></div>`);
		}
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
//#region src/lib/modules/TaskBoard.svelte
function TaskBoard($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let roundState = null;
		function getActiveSlot(rs) {
			if (!rs) return null;
			return rs.slots.find((s) => s.name === rs.current_slot) ?? null;
		}
		const activeSlot = derived(() => getActiveSlot(roundState));
		$$renderer.push(`<div class="board svelte-figrcc">`);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <div${attr_class("chain svelte-figrcc", void 0, { "paused": roundState?.pause_active })}>`);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<div class="loading svelte-figrcc">Syncing round state…</div>`);
		$$renderer.push(`<!--]--></div> `);
		if (activeSlot() && roundState);
		else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
//#region src/lib/modules/SystemLog.svelte
function SystemLog($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let members = [];
		let selectedId = "";
		let lines = {};
		let paneAvailable = false;
		const currentLines = derived(() => lines[selectedId] ?? []);
		const displayMembers = derived(() => members.length > 0 ? members : agentStore.agents.map((a) => ({
			id: a.id,
			name: a.name,
			fingerprint: "",
			bindrune_url: "",
			status: a.status === "working" ? "working" : "idle",
			task: a.task,
			pane_cmd: "",
			last_seen: a.last_seen,
			source: "sygil"
		})));
		const lineTypeClass = {
			stdout: "",
			status: "line-status",
			error: "line-error",
			info: "line-info"
		};
		$$renderer.push(`<div class="syslog svelte-fo3dsi"><div class="tabs svelte-fo3dsi"><!--[-->`);
		const each_array = ensure_array_like(displayMembers());
		for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
			let m = each_array[$$index];
			$$renderer.push(`<button${attr_class("tab svelte-fo3dsi", void 0, { "active": selectedId === m.id })}>${escape_html(m.name)}</button>`);
		}
		$$renderer.push(`<!--]--> `);
		if (displayMembers().length === 0) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<span class="no-agents svelte-fo3dsi">No agents registered</span>`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <div class="tab-spacer svelte-fo3dsi"></div> <span${attr_class("source-indicator svelte-fo3dsi", void 0, { "live": paneAvailable })}>${escape_html("○ polled")}</span></div> <div class="output svelte-fo3dsi">`);
		if (currentLines().length === 0) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<div class="empty-log svelte-fo3dsi">${escape_html("Select an agent above")}</div>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--[-->`);
			const each_array_1 = ensure_array_like(currentLines());
			for (let i = 0, $$length = each_array_1.length; i < $$length; i++) {
				let line = each_array_1[i];
				$$renderer.push(`<div${attr_class(`line ${stringify(lineTypeClass[line.type])}`, "svelte-fo3dsi")}><span class="line-ts svelte-fo3dsi">${escape_html(line.ts)}</span> <span class="line-text svelte-fo3dsi">${escape_html(line.text)}</span></div>`);
			}
			$$renderer.push(`<!--]-->`);
		}
		$$renderer.push(`<!--]--> <div></div></div> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
//#region src/lib/modules/ResourceMeter.svelte
function ResourceMeter($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		$$renderer.push(`<div class="meter svelte-1nkmd72">`);
		{
			$$renderer.push("<!--[1-->");
			$$renderer.push(`<!--[-->`);
			const each_array = ensure_array_like(Array(2));
			for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
				each_array[$$index];
				$$renderer.push(`<div class="host-skeleton svelte-1nkmd72"><div class="skel-label svelte-1nkmd72"></div> <div class="skel-bar svelte-1nkmd72"></div> <div class="skel-bar short svelte-1nkmd72"></div></div>`);
			}
			$$renderer.push(`<!--]-->`);
		}
		$$renderer.push(`<!--]--></div>`);
	});
}
//#endregion
//#region src/routes/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		const MODULE_COMPONENTS = {
			hub: AgentHub,
			chat: Chat,
			canvas: Canvas,
			presence: CohortPresence,
			stream: ConversationStream,
			decisions: DecisionLog,
			taskboard: TaskBoard,
			syslog: SystemLog,
			resources: ResourceMeter
		};
		$$renderer.push(`<div class="shell svelte-1uha8ag"><div class="grid-bg svelte-1uha8ag"></div> <!--[-->`);
		const each_array = ensure_array_like(boxStore.boxes);
		for (let i = 0, $$length = each_array.length; i < $$length; i++) {
			let box = each_array[i];
			const Comp = MODULE_COMPONENTS[box.module];
			Box($$renderer, {
				box,
				zIndex: 100 + i,
				children: ($$renderer) => {
					if (Comp) {
						$$renderer.push("<!--[0-->");
						if (Comp) {
							$$renderer.push("<!--[-->");
							Comp($$renderer, {});
							$$renderer.push("<!--]-->");
						} else {
							$$renderer.push("<!--[!-->");
							$$renderer.push("<!--]-->");
						}
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
