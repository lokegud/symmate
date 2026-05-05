export type ModuleType = 'hub' | 'chat' | 'canvas';

export interface BoxState {
	id: string;
	module: ModuleType;
	title: string;
	tint: string;         // CSS color
	x: number;
	y: number;
	w: number;
	h: number;
	// module-specific config
	agentId?: string;     // for chat module
}

const DEFAULTS: BoxState[] = [
	{ id: 'hub',    module: 'hub',    title: 'Agents',  tint: '#3d8ef0', x: 20,  y: 20,  w: 400, h: 500 },
	{ id: 'canvas', module: 'canvas', title: 'Canvas',  tint: '#10b981', x: 440, y: 20,  w: 500, h: 500 },
	{ id: 'chat',   module: 'chat',   title: 'Chat',    tint: '#8b5cf6', x: 20,  y: 540, w: 920, h: 340 },
];

function loadBoxes(): BoxState[] {
	try {
		const saved = localStorage.getItem('conclave:boxes');
		if (saved) return JSON.parse(saved);
	} catch {}
	return DEFAULTS;
}

function save(boxes: BoxState[]) {
	try { localStorage.setItem('conclave:boxes', JSON.stringify(boxes)); } catch {}
}

class BoxStore {
	boxes = $state<BoxState[]>(loadBoxes());

	update(id: string, patch: Partial<BoxState>) {
		this.boxes = this.boxes.map(b => b.id === id ? { ...b, ...patch } : b);
		save(this.boxes);
	}

	add(b: BoxState) {
		this.boxes = [...this.boxes, b];
		save(this.boxes);
	}

	remove(id: string) {
		this.boxes = this.boxes.filter(b => b.id !== id);
		save(this.boxes);
	}

	bringToFront(id: string) {
		const box = this.boxes.find(b => b.id === id);
		if (!box) return;
		this.boxes = [...this.boxes.filter(b => b.id !== id), box];
		save(this.boxes);
	}
}

export const boxStore = new BoxStore();

// Agent data from hub server SSE
export interface Agent {
	id: string;
	name: string;
	task: string;
	status: 'idle' | 'working' | 'blocked' | 'done';
	todos: { id: string; text: string; done: boolean }[];
	chat_url: string;
	last_seen: string;
}

class AgentStore {
	agents = $state<Agent[]>([]);

	set(agents: Agent[]) {
		this.agents = agents;
	}
}

export const agentStore = new AgentStore();

export const HUB_URL = '/api';
