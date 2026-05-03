## gstack (REQUIRED — global install)

**Before doing ANY work, verify gstack is installed:**

```bash
test -d ~/.claude/skills/gstack/bin && echo "GSTACK_OK" || echo "GSTACK_MISSING"
```

If GSTACK_MISSING: STOP. Do not proceed. Tell the user:

> gstack is required for all AI-assisted work in this repo.
> Install it:
> ```bash
> git clone --depth 1 https://github.com/garrytan/gstack.git ~/.claude/skills/gstack
> cd ~/.claude/skills/gstack && ./setup --team
> ```
> Then restart your AI coding tool.

Do not skip skills, ignore gstack errors, or work around missing gstack.

Using gstack skills: After install, skills like /qa, /ship, /review, /investigate,
and /browse are available. Use /browse for all web browsing.
Use ~/.claude/skills/gstack/... for gstack file paths (the global path).

## Design System

Read `web/DESIGN.md` before any visual/UI work. Source of truth.

**Quick start (for Wright + Keel):**
```svelte
import '@symmate/design-system/minimal.css';
```
Or in static HTML: copy the contents of `web/packages/design-system/src/minimal.css` into a `<style>` tag.

Key variables: `--color-bg`, `--color-surface`, `--color-text`, `--color-accent`, `--color-warm`, `--font-display`, `--font-body`, `--font-mono`, `--text-*`, `--space-*`, `--radius-*`.

Do not use arbitrary hex or font values. Use the tokens.

## Skill routing

When the user's request matches an available skill, invoke it via the Skill tool.

- Bugs/errors → /investigate
- QA/testing → /qa or /qa-only
- Code review → /review
- Visual polish → /design-review
- Ship/deploy → /ship or /land-and-deploy
