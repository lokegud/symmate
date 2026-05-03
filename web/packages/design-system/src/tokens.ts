/** Symmate Design System — TypeScript token exports
 * Wren · 2026-05-02
 *
 * These mirror tokens.css and are the authoritative typed values.
 * Import for use in Svelte component logic, Tailwind config, or testing.
 */

export const fonts = {
  display: "'Fraunces', Georgia, serif",
  body:    "'Plus Jakarta Sans', system-ui, sans-serif",
  mono:    "'JetBrains Mono', 'Fira Code', monospace",
} as const;

export const textSizes = {
  xs:   '0.6875rem',
  sm:   '0.8125rem',
  base: '1rem',
  md:   '1.125rem',
  lg:   '1.25rem',
  xl:   '1.5rem',
  '2xl': '1.875rem',
  '3xl': '2.25rem',
  '4xl': '3rem',
  '5xl': '4rem',
  '6xl': '5.5rem',
} as const;

export const colors = {
  light: {
    bg:             '#F5F2ED',
    surface:        '#FFFFFF',
    surfaceRaised:  '#F0EDE7',
    text:           '#1C1A16',
    textMuted:      '#6B6862',
    textSubtle:     '#9B9890',
    textInverse:    '#F5F2ED',
    border:         '#E0DCD5',
    borderSubtle:   '#EAE7E1',
    borderStrong:   '#C8C4BC',
    accent:         '#4B5FD3',
    accentHover:    '#3B50C8',
    accentPressed:  '#2E42BA',
    accentSubtle:   '#EEF0FA',
    warm:           '#D4895A',
    warmHover:      '#C47848',
    warmSubtle:     '#FAF0EB',
    prismStart:     '#4B5FD3',
    prismMid:       '#8B5BD3',
    prismEnd:       '#D4895A',
    success:        '#38956A',
    successSubtle:  '#EBF6F1',
    warning:        '#D48A2A',
    warningSubtle:  '#FDF3E4',
    error:          '#C94444',
    errorSubtle:    '#FAEAEA',
    info:           '#4B5FD3',
    infoSubtle:     '#EEF0FA',
  },
  dark: {
    bg:             '#0E0C0A',
    surface:        '#161310',
    surfaceRaised:  '#1E1B17',
    text:           '#EDE9E2',
    textMuted:      '#8C8880',
    textSubtle:     '#5C5850',
    textInverse:    '#1C1A16',
    border:         '#282420',
    borderSubtle:   '#201D19',
    borderStrong:   '#383430',
    accent:         '#6C7EE8',
    accentHover:    '#7D8EEF',
    accentPressed:  '#8C9CF2',
    accentSubtle:   '#1A1E38',
    warm:           '#E8A878',
    warmHover:      '#F0B888',
    warmSubtle:     '#2A1E14',
    success:        '#52B882',
    successSubtle:  '#0E2018',
    warning:        '#E8A83A',
    warningSubtle:  '#281A08',
    error:          '#E06060',
    errorSubtle:    '#2A0E0E',
  },
} as const;

export const spacing = {
  '0':    '0',
  px:     '1px',
  '0.5':  '2px',
  '1':    '4px',
  '1.5':  '6px',
  '2':    '8px',
  '3':    '12px',
  '4':    '16px',
  '5':    '20px',
  '6':    '24px',
  '8':    '32px',
  '10':   '40px',
  '12':   '48px',
  '16':   '64px',
  '20':   '80px',
  '24':   '96px',
  '32':   '128px',
} as const;

export const radii = {
  sm:   '6px',
  md:   '10px',
  lg:   '16px',
  xl:   '24px',
  '2xl':'32px',
  full: '9999px',
} as const;

export const shadows = {
  sm:        '0 1px 2px rgba(0,0,0,0.06), 0 1px 1px rgba(0,0,0,0.04)',
  md:        '0 4px 8px rgba(0,0,0,0.08), 0 2px 4px rgba(0,0,0,0.04)',
  lg:        '0 8px 24px rgba(0,0,0,0.10), 0 4px 8px rgba(0,0,0,0.06)',
  xl:        '0 16px 48px rgba(0,0,0,0.12), 0 8px 16px rgba(0,0,0,0.06)',
  prism:     '0 0 32px rgba(75,95,211,0.20), 0 0 64px rgba(139,91,211,0.12)',
  prismWarm: '0 0 24px rgba(212,137,90,0.25)',
} as const;

export const durations = {
  instant:  '50ms',
  micro:    '100ms',
  fast:     '150ms',
  normal:   '220ms',
  moderate: '350ms',
  slow:     '500ms',
  blob:     '800ms',
  pulse:    '3000ms',
} as const;

export const easings = {
  out:    'cubic-bezier(0.22, 1, 0.36, 1)',
  in:     'cubic-bezier(0.5, 0, 0.75, 0)',
  inout:  'cubic-bezier(0.4, 0, 0.2, 1)',
  spring: 'cubic-bezier(0.34, 1.4, 0.64, 1)',
  back:   'cubic-bezier(0.34, 1.56, 0.64, 1)',
} as const;

export const breakpoints = {
  sm:  480,
  md:  768,
  lg:  1024,
  xl:  1280,
  '2xl': 1536,
} as const;

export const widths = {
  content:   '720px',
  platform:  '1200px',
  dashboard: '1440px',
} as const;

export type BlobState =
  | 'idle' | 'thinking' | 'listening' | 'happy'
  | 'content' | 'nervous' | 'reluctant' | 'guarding'
  | 'disagreeing' | 'dormant';

export type BlobMode = 'per' | 'bus';
export type BlobSize = 'sm' | 'md' | 'lg' | 'xl';
