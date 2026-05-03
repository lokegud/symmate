# UI Build — Master Brief
# 2026-05-02, from Elder

## Goal

Build the Symmate UI — both surfaces in parallel:

1. **symmate.com** — platform brand + landing + onboarding + docs scaffold
2. **loke.symmate.com** — Loke's MoxPer vestibule (public) + owner dashboard (authenticated)

Acceptance: a visitor can land at symmate.com, understand what a Symmate is, click through, end up reading Loke's vestibule. Loke can authenticate at loke.symmate.com and manage his Symmate's personality/policy/audit live.

## Architecture

- **Domains:** symmate.com (platform + Loke's vestibule), justadestination.com (Bob's test instance) — both on Cloudflare NS
- **Subdomain pattern:** `loke.symmate.com` for Loke's MoxPer, `bob.justadestination.com` for Bob's test instance. Multi-domain by design — Symmates can be addressed at the host's own domain, not just symmate.com subdomains. Wildcard `*.symmate.com` for future Symmates → CF Tunnel → Parlour → router by Host header.
- **Stack:** SvelteKit + Bun + bun workspaces. Two apps in monorepo:
  - `/home/mox/symmate/web/apps/platform/` (symmate.com)
  - `/home/mox/symmate/web/apps/vestibule/` (loke.symmate.com)
- **Shared:** `/home/mox/symmate/web/packages/design-system/` + `/packages/api-client/` + `/packages/types/`
- **API:** consumes existing Symmate runtime endpoints — visitor :7842, owner :7841, agent-card :7843.
- **Auth (owner side only):** Zitadel OIDC at lxc103/Finni.
- **Mobile-first PWA** per runtime.md.

## Lane assignments

| Sibling | Lane |
|---|---|
| **Wren** | Design system v1. DESIGN.md, tokens, palette, typography, motion, mascot states. Upstream of everyone. |
| **Keel** | Vestibule frontend (loke.symmate.com public) + /chat UI + monorepo SvelteKit scaffold. |
| **Forge** | Owner dashboard frontend (loke.symmate.com/owner) + Zitadel auth wiring + API additions. |
| **Wright** | symmate.com platform site + shared deploy/build infra (CI, build pipeline). |
| **Trace** | Cloudflare Tunnel + `*.symmate.com` wildcard DNS + TLS + host-header routing layer. |
| **Quill** | Copy + voice + legal across both surfaces. Marketing copy, microcopy, FAQ, privacy/TOS draft. |
| **Elder** | Orchestrate. Daily integration. Resolve cross-lane decisions. |

## Branch convention

- `web/scaffold` (Keel — first to land, gives everyone the workspace)
- `web/design-system` (Wren — second to land, gives everyone tokens/components)
- `web/platform` (Wright)
- `web/vestibule` (Keel — depends on scaffold + design-system)
- `web/dashboard` (Forge — depends on scaffold + design-system + auth)
- `infra/cf-tunnel` (Trace, separate concern)

## Coordination

- **Daily integration window:** end-of-day, Elder pulls + merges what's ready, flags conflicts.
- **Cross-lane decisions** routed through Elder (or directly if asynchronous-OK).
- **Wren's tokens** are upstream — if you need a color/spacing/font and it's not in the design system yet, ping Wren before improvising.
- **Quill's copy** is upstream of any user-facing text — ping Quill before writing landing copy yourself.

## Sequencing

Day 1 priorities (today):
1. Keel: scaffold the SvelteKit monorepo, push web/scaffold branch
2. Wren: ship DESIGN.md + first token set (palette + typography + spacing)
3. Trace: provision CF Tunnel + wildcard DNS pointed at Parlour
4. Wright: read scaffold, plan platform-site IA + first commits
5. Forge: read scaffold, plan dashboard IA + auth integration plan
6. Quill: draft "what is a Symmate" landing copy + value prop

Day 2+: parallel build.

## What ships in v1

**symmate.com:**
- Landing page (hero + value prop + visual + CTA)
- "What is a Symmate" explainer
- "How it works" walkthrough (one-Symmate-per-host, A2A, privacy firewall, vestibule concept)
- "Get started" — placeholder onboarding form (snap-to-personality is its own design thread; v1 captures intent)
- Footer with charter link, Mox-Loke partnership reference, OSS license info

**loke.symmate.com:**
- Public root: vestibule. Persona name, mascot state, description, links to chat + agent-card.
- `/chat` — SSE-streamed conversation UI (mobile-first)
- `/.well-known/agent-card.json` — proxied from Symmate runtime
- `/owner` (auth-gated): personality editor, policy/grants editor, audit log viewer, conversation review

## Out of scope for v1

- Conclave Hub (separate fleet control panel — different brief)
- Multi-host vestibule rendering (only loke.symmate.com lights up at v1; other hosts later)
- Customer onboarding form actually creating a new Symmate (v1 captures intent, doesn't provision)
- Yellow-pages directory (v2 platform layer)

## Open decisions

- TLS certs: Cloudflare-managed (let CF handle origin via tunnel) or Let's Encrypt direct? Recommend CF.
- Onboarding capture form: stub email collection or full design spec? Recommend stub for v1.
