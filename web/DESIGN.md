# Design System — Symmate
**Status:** `standard` (v1.0 — Wren, 2026-05-02)

---

## Product Context

- **What this is:** A privacy-first personal AI representative — one per person, one per business, speaking for you in an A2A world full of vendor agents.
- **Who it's for:** Individuals who want a buddy with standing to push back; SMBs who want a representative that doesn't embarrass them.
- **Space:** Consumer-side AI sovereignty. The counterweight to the 150 enterprise agents already running.
- **Surfaces:** `symmate.com` (brand + platform landing) · `loke.symmate.com/` (vestibule) · `/chat` (SSE streaming) · `/owner` (auth-gated dashboard)
- **Stack:** SvelteKit 5 + Bun · mobile-first PWA

**Memorable thing:** *"Like a Tamagotchi that will fix your website — and might tell you not to send that email."*

Every design decision should serve this: warmth + earned friction. Not a tool with a face. Someone.

---

## Aesthetic Direction

**Direction: Luminous Organic**

Translucent surfaces. Iridescent accents. Soft-edged forms that breathe. The UI is made of the same material as the Blob — semi-transparent, lit from within, warm at rest and sharp under pressure.

**What it is not:**
- Not glassmorphism excess (no frosted blur on everything)
- Not flat minimal (there is depth and light)
- Not enterprise gray (warmth is structural, not decorative)
- Not startup-purple-gradient (the prism is earned, not default)

**Decoration level:** Intentional — subtle luminosity and surface texture at key moments. Typography carries the hierarchy; decoration punctuates.

**Mood:** Calm confidence. A warm room with very good lighting. The kind of software that feels like it has already thought about your problem before you finished describing it.

**Reference energy:** Linear meets Framer meets a really good field notes journal. Precision tools that feel handmade.

---

## Typography

### Stack

| Role | Font | Why |
|---|---|---|
| **Display / Hero** | [Fraunces](https://fonts.google.com/specimen/Fraunces) | Variable optical serif — personality, warmth, character. Unique in tech; pairs with the organic-blob identity. The serif says "I'm someone, not something." |
| **Body / UI** | [Plus Jakarta Sans](https://fonts.google.com/specimen/Plus+Jakarta+Sans) | Friendly, large x-height, excellent mobile legibility. Warm grotesque without being round to the point of childishness. |
| **Mono** | [JetBrains Mono](https://fonts.google.com/specimen/JetBrains+Mono) | Audit logs, agent cards, code blocks, API keys. Ligature support. |

### Loading

```html
<!-- In app.html <head> -->
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Fraunces:ital,opsz,wght@0,9..144,300..900;1,9..144,300..900&family=Plus+Jakarta+Sans:wght@300..800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

Self-host for production via `fontsource` packages (avoids Google Fonts blocking in privacy-conscious contexts).

### Type Scale

```css
/* All in rem — base 16px */
--text-xs:    0.6875rem;   /* 11px — legal, timestamps */
--text-sm:    0.8125rem;   /* 13px — metadata, labels */
--text-base:  1rem;        /* 16px — body */
--text-md:    1.125rem;    /* 18px — emphasized body */
--text-lg:    1.25rem;     /* 20px — card titles */
--text-xl:    1.5rem;      /* 24px — section headers */
--text-2xl:   1.875rem;    /* 30px — page headers */
--text-3xl:   2.25rem;     /* 36px — display */
--text-4xl:   3rem;        /* 48px — hero (mobile) */
--text-5xl:   4rem;        /* 64px — hero (desktop) */
--text-6xl:   5.5rem;      /* 88px — hero max */
```

### Font Weights

```css
--weight-light:    300;
--weight-regular:  400;
--weight-medium:   500;
--weight-semibold: 600;
--weight-bold:     700;
--weight-black:    800;
```

**Fraunces usage:** Display only — hero headings, marketing section titles, the vestibule persona name. Use the optical axis: `font-optical-sizing: auto`. Italic for emphasis in editorial contexts.

**Plus Jakarta Sans usage:** Everything else — nav, body, UI labels, buttons, form inputs.

**Leading:**
```css
--leading-tight:  1.15;  /* display text */
--leading-snug:   1.3;   /* headings */
--leading-normal: 1.55;  /* body */
--leading-loose:  1.75;  /* long-form readable content */
```

**Tracking:**
```css
--tracking-tight:  -0.03em;  /* large display */
--tracking-normal:  0;
--tracking-wide:    0.04em;  /* small caps, labels, badges */
--tracking-widest:  0.1em;   /* legal, timestamps */
```

---

## Color

### Approach: Prismatic Warm

Three accent nodes derived from the Blob's body gradient — blue crown → violet midpoint → amber midsection. Regular UI uses blue as primary. The full prism is reserved for brand moments, mascot states, and the onboarding ritual.

### Light Mode

```css
/* Surfaces */
--color-bg:              #F5F2ED;  /* Warm vellum — not clinical white */
--color-surface:         #FFFFFF;
--color-surface-raised:  #F0EDE7;  /* Elevated cards */
--color-surface-overlay: rgba(255, 255, 255, 0.85); /* Glass panels */

/* Text */
--color-text:            #1C1A16;  /* Warm near-black */
--color-text-muted:      #6B6862;  /* Secondary content */
--color-text-subtle:     #9B9890;  /* Placeholders, disabled */
--color-text-inverse:    #F5F2ED;  /* On dark/accent surfaces */

/* Borders */
--color-border:          #E0DCD5;
--color-border-subtle:   #EAE7E1;
--color-border-strong:   #C8C4BC;

/* Accent — primary (Blob crown blue) */
--color-accent:          #4B5FD3;
--color-accent-hover:    #3B50C8;
--color-accent-pressed:  #2E42BA;
--color-accent-subtle:   #EEF0FA;
--color-accent-text:     #FFFFFF;

/* Accent — secondary (Blob midsection amber) */
--color-warm:            #D4895A;
--color-warm-hover:      #C47848;
--color-warm-subtle:     #FAF0EB;
--color-warm-text:       #FFFFFF;

/* Prism (gradient accent — brand moments only) */
--color-prism-start:     #4B5FD3;  /* Blue */
--color-prism-mid:       #8B5BD3;  /* Violet */
--color-prism-end:       #D4895A;  /* Amber */
--gradient-prism: linear-gradient(135deg, var(--color-prism-start), var(--color-prism-mid), var(--color-prism-end));
--gradient-prism-subtle: linear-gradient(135deg,
  color-mix(in srgb, var(--color-prism-start) 15%, transparent),
  color-mix(in srgb, var(--color-prism-mid) 15%, transparent),
  color-mix(in srgb, var(--color-prism-end) 15%, transparent));

/* Semantic */
--color-success:         #38956A;
--color-success-subtle:  #EBF6F1;
--color-warning:         #D48A2A;
--color-warning-subtle:  #FDF3E4;
--color-error:           #C94444;
--color-error-subtle:    #FAEAEA;
--color-info:            #4B5FD3;   /* same as accent */
--color-info-subtle:     #EEF0FA;
```

### Dark Mode

```css
[data-theme="dark"] {
  /* Surfaces */
  --color-bg:              #0E0C0A;  /* Very warm black */
  --color-surface:         #161310;
  --color-surface-raised:  #1E1B17;
  --color-surface-overlay: rgba(22, 19, 16, 0.90);

  /* Text */
  --color-text:            #EDE9E2;
  --color-text-muted:      #8C8880;
  --color-text-subtle:     #5C5850;
  --color-text-inverse:    #1C1A16;

  /* Borders */
  --color-border:          #282420;
  --color-border-subtle:   #201D19;
  --color-border-strong:   #383430;

  /* Accent */
  --color-accent:          #6C7EE8;
  --color-accent-hover:    #7D8EEF;
  --color-accent-pressed:  #8C9CF2;
  --color-accent-subtle:   #1A1E38;

  /* Warm */
  --color-warm:            #E8A878;
  --color-warm-hover:      #F0B888;
  --color-warm-subtle:     #2A1E14;

  /* Semantic */
  --color-success:         #52B882;
  --color-success-subtle:  #0E2018;
  --color-warning:         #E8A83A;
  --color-warning-subtle:  #281A08;
  --color-error:           #E06060;
  --color-error-subtle:    #2A0E0E;
}
```

### Contrast requirements

- Body text on bg: 7:1+ (AAA)
- Muted text on bg: 4.5:1+ (AA)
- All interactive states: visible at 3:1+ against adjacent color
- Focus rings: `--color-accent` at 2px offset, 2px width

---

## Spacing

**Base unit: 4px.** Comfortable density — not data-dense like a dashboard, not airy like a pure marketing site.

```css
--space-0:   0;
--space-px:  1px;
--space-0-5: 2px;
--space-1:   4px;
--space-1-5: 6px;
--space-2:   8px;
--space-3:   12px;
--space-4:   16px;
--space-5:   20px;
--space-6:   24px;
--space-8:   32px;
--space-10:  40px;
--space-12:  48px;
--space-16:  64px;
--space-20:  80px;
--space-24:  96px;
--space-32:  128px;
```

**Semantic aliases:**
```css
--inset-xs:  var(--space-2);   /* Tight component padding */
--inset-sm:  var(--space-3);   /* Compact padding */
--inset-md:  var(--space-4);   /* Default component padding */
--inset-lg:  var(--space-6);   /* Generous component padding */
--inset-xl:  var(--space-8);   /* Section padding */
--stack-xs:  var(--space-1);   /* Tight vertical rhythm */
--stack-sm:  var(--space-2);   /* List items */
--stack-md:  var(--space-4);   /* Component stacks */
--stack-lg:  var(--space-6);   /* Section stacks */
--stack-xl:  var(--space-12);  /* Page sections */
```

---

## Layout + Grid

```css
/* Grid */
--grid-cols-mobile:  4;
--grid-cols-tablet:  8;
--grid-cols-desktop: 12;
--gutter-mobile:     16px;
--gutter-tablet:     24px;
--gutter-desktop:    32px;

/* Breakpoints */
--bp-sm:  480px;
--bp-md:  768px;
--bp-lg:  1024px;
--bp-xl:  1280px;
--bp-2xl: 1536px;

/* Content widths */
--width-content:   720px;   /* Long-form content, chat */
--width-platform:  1200px;  /* symmate.com platform */
--width-dashboard: 1440px;  /* Owner dashboard, data-heavy */

/* Border radius */
--radius-sm:   6px;    /* Badges, chips, small tags */
--radius-md:   10px;   /* Buttons, inputs, small cards */
--radius-lg:   16px;   /* Cards, panels */
--radius-xl:   24px;   /* Modals, drawers, large cards */
--radius-2xl:  32px;   /* Hero cards, feature blocks */
--radius-full: 9999px; /* Pills, avatars */

/* Blob radius — varies by state, handled in MascotBlob component */
/* Base idle shape approximated with: */
--radius-blob: 60% 40% 55% 45% / 50% 45% 55% 50%;
```

### Surface elevation

```css
--shadow-sm:  0 1px 2px rgba(0,0,0,0.06), 0 1px 1px rgba(0,0,0,0.04);
--shadow-md:  0 4px 8px rgba(0,0,0,0.08), 0 2px 4px rgba(0,0,0,0.04);
--shadow-lg:  0 8px 24px rgba(0,0,0,0.10), 0 4px 8px rgba(0,0,0,0.06);
--shadow-xl:  0 16px 48px rgba(0,0,0,0.12), 0 8px 16px rgba(0,0,0,0.06);

/* Prism glow — blob and brand moments */
--shadow-prism: 0 0 32px rgba(75, 95, 211, 0.20), 0 0 64px rgba(139, 91, 211, 0.12);
--shadow-prism-warm: 0 0 24px rgba(212, 137, 90, 0.25);
```

---

## Motion

**Approach: Intentional** — animate what aids comprehension, choreograph what encodes personality. The Blob breathes; the UI doesn't fight it.

### Timing functions

```css
/* Entry: exponential ease-out — snappy arrival */
--ease-out:   cubic-bezier(0.22, 1, 0.36, 1);

/* Exit: ease-in — things that leave don't linger */
--ease-in:    cubic-bezier(0.5, 0, 0.75, 0);

/* Move: balanced — state changes, repositioning */
--ease-inout: cubic-bezier(0.4, 0, 0.2, 1);

/* Spring: for blob and organic elements */
--ease-spring: cubic-bezier(0.34, 1.4, 0.64, 1);

/* Anticipation: slight overshoot for playful moments */
--ease-back:  cubic-bezier(0.34, 1.56, 0.64, 1);
```

### Duration scale

```css
--duration-instant:  50ms;   /* Instant feedback (color changes) */
--duration-micro:    100ms;  /* Hover states, focus rings */
--duration-fast:     150ms;  /* Exits, quick state changes */
--duration-normal:   220ms;  /* Standard entries, most transitions */
--duration-moderate: 350ms;  /* Page sections, larger cards */
--duration-slow:     500ms;  /* Modals, drawers, page transitions */
--duration-blob:     800ms;  /* Blob morph transitions — spring */
--duration-pulse:   3000ms;  /* Blob idle breathing cycle */
```

### When to animate

**Always animate:** focus rings, hover states, active/pressed, toast entries, modal open/close, chat bubble entry, loading states.

**Selectively animate:** page section reveals (only on first view), mascot state changes, audit log entries, typing indicator.

**Never animate:** form validation errors (must appear instantly for usability), table sort changes, content edits in-place.

### Blob animation spec

The Blob's idle breathing is a CSS `border-radius` keyframe animation at `--duration-pulse` with `--ease-inout`. Morph states (idle → thinking, thinking → speaking, etc.) use a 2-keyframe transition at `--duration-blob` with `--ease-spring`. Expression overlays (eyes, smile curve) are SVG `<animate>` or WAAPI with spring easing.

The prism gradient shimmer: `@keyframes prism-shift` rotates `background-position` on a 200% background-size gradient over 4s `infinite alternate ease-in-out`. Used only on the Blob and brand marks — never on generic UI.

### Reduced motion

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
  /* Exception: the Blob still performs its state change (accessibility-meaningful)
     but the morph is cut-to instead of animated */
}
```

---

## Component Catalog

### Primitive components (v1 — ship in design-system package)

| Component | Status | Notes |
|---|---|---|
| `Button` | `standard` | Variants: primary, secondary, ghost, danger, icon |
| `Input` | `standard` | text, textarea, with label, hint, error state |
| `Card` | `standard` | elevated, outlined, flat; composable header/body/footer |
| `Badge` | `standard` | semantic: success/warning/error/info/neutral + custom |
| `Avatar` | `standard` | image with fallback, size scale, ring variant |
| `MascotBlob` | `candidate` | Generative sigil — seed + state + evolved flag |
| `ChatBubble` | `candidate` | user/assistant variants, streaming state |
| `Spinner` | `standard` | single ring, inherits accent color |
| `Divider` | `standard` | horizontal and vertical, with optional label |

### App-level components (built in apps, use design-system primitives)

- `VestibuleHero` — persona name (Fraunces display), mascot state, description, links
- `ChatPane` — scrolling bubble list + composer + SSE streaming indicator
- `OwnerNav` — authenticated sidebar nav for /owner dashboard
- `PersonalityEditor` — structured form for personality/policy editing
- `AuditLog` — paginated list of AuditEntry rows
- `StatCard` — metric with label, trend, icon

---

## Mascot Integration

### The Blob — design invariants

These survive across every morph state and expression (from `mox-edu-mascot.md`):

1. **The eyes** — large round pupils, bright white sparkle highlight, warm gaze. Identity anchor.
2. **Rainbow-iridescent gradient** — `--color-prism-start` at crown through `--color-prism-mid` to `--color-prism-end` at midsection. This gradient IS the creature.
3. **Internal drifting marks** — visible through a semi-transparent body. Floating, not static. The creature is made of something translucent.
4. **The smile** — present in all non-friction expressions. Small, warm, not wide.
5. **Palm-sized scale** — always reads as small-companion. Never imposing.

### Expression → UI state mapping

| Expression | When it fires |
|---|---|
| `idle` (neutral/curious) | Default; waiting for input |
| `thinking` | Processing / loading / SSE streaming |
| `listening` | User is typing / speaking |
| `happy` | Task completed, message sent |
| `content` | Session end, calm moment |
| `nervous` | High-stakes action pending, error approaching |
| `reluctant` | Evaluating whether to comply with a request |
| `guarding` | Intercepting / blocking / "wait" |
| `disagreeing` | Principled refusal |
| `dormant` | Inactive, session idle >5min |

**Mox.Per (personal) vs Mox.Bus (business):** Same Blob species. Personal variant fires `reluctant` / `guarding` / `disagreeing` freely (the Pikachu-shocks-Ash relationship). Business variant fires friction expressions rarely — fraud, abuse, scope violation only. The `seed` prop differentiates appearance; the `mode` prop calibrates friction expression frequency.

### MascotBlob component API

```typescript
interface MascotBlobProps {
  seed: string;        // Per-Symmate personality seed (determines marking pattern)
  state: BlobState;    // Expression state (see above)
  size?: 'sm' | 'md' | 'lg' | 'xl';  // Default: 'md'
  evolved?: boolean;   // false = sparse marks (new instance), true = dense (seasoned)
  mode?: 'per' | 'bus'; // Friction calibration — default 'per'
  animated?: boolean;  // Disable for reduced-motion contexts
}

type BlobState =
  | 'idle' | 'thinking' | 'listening' | 'happy'
  | 'content' | 'nervous' | 'reluctant' | 'guarding'
  | 'disagreeing' | 'dormant';
```

**v1 implementation:** Static state images (PNG/SVG per expression) with CSS transition between states. The `seed` seeds a deterministic mark-density and color-temperature calculation that selects from a pre-rendered variant set.

**v2 target:** Procedural CSS + canvas — the Blob is drawn programmatically from the seed, marks are generated and drift on a canvas layer, morph states use WAAPI spring transitions.

### Platform vs per-Symmate mascot

- `symmate.com` uses the canonical platform mascot (seed: `platform-canonical`) — neutral gradient, balanced marks, `idle` and `happy` states only
- `loke.symmate.com` uses Loke's seeded variant — personality-specific mark density, full expression palette available

### Marking language (prototype status)

| Mark quality | Signals |
|---|---|
| Dot density | Sociability — dense = extroverted, sparse = reserved |
| Color temperature | Emotional baseline — warm (amber) = optimistic, cool (blue) = analytical |
| Mark luminosity | Activation — bright/fast = processing, dim/slow = resting |
| Clustering location | Priority axis — upper body = ideas/comms, lower body = action |

Marks evolve over time. A new instance has sparse generic marks. A seasoned instance has a dense legible pattern. This is the visual evolution arc — the trust-building mechanic made visible.

---

## Decisions Log

| Date | Decision | Rationale |
|---|---|---|
| 2026-05-02 | Fraunces as display type | Unique in tech; personality-forward; warmth pairs with the Blob's organic identity. Avoids the grotesque monoculture (Inter, Space Grotesk). |
| 2026-05-02 | Plus Jakarta Sans as body/UI | Warm grotesque, excellent mobile legibility, good variable weight range. Not overused. |
| 2026-05-02 | Warm vellum background (#F5F2ED) over clinical white | Signals "alive, inhabited" rather than "empty form". The warmth is structural. |
| 2026-05-02 | Prismatic three-node accent (blue → violet → amber) | Derived from Blob's gradient — brand and mascot visually unified. Reserved for brand moments to avoid cheapening. |
| 2026-05-02 | 4px base spacing unit, comfortable density | Between dashboard-compact and marketing-airy. The vestibule and platform are both used, not just read. |
| 2026-05-02 | Border-radius hierarchy (6/10/16/24/32) | Avoids the uniform-bubble anti-pattern. Larger containers get more rounding; small components stay crisp. |
| 2026-05-02 | Spring easing for Blob, standard ease-out for UI | Blob has its own physics. UI doesn't compete. |
| 2026-05-02 | Luminous Organic aesthetic direction | Translucency + iridescence + warmth — the UI is made of the same material as the Blob. Not glassmorphism excess. Not flat minimal. |
| 2026-05-02 | Dark mode: very warm black (#0E0C0A) | Clinical dark (#000 or #111) would fight the warm brand. The warmth is consistent across light and dark. |

---

*Authored by Wren (Mox.Design) · `web/design-system` branch · 2026-05-02*
*Artifact class: `standard` (v1.0) per WREN_CHARTER.md §A1*
*Upstream of: Keel (vestibule), Forge (dashboard), Wright (platform)*
