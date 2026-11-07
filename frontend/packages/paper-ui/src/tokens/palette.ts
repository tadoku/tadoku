/**
 * Primitive colours. Components and pages never use these directly: semantic colours in `theme.ts` pick from them
 * per theme, so a palette change moves every surface that shares a role.
 */
export const palette = {
  white: '#ffffff',
  /** Warm paper neutrals, light theme. */
  sand: {
    50: '#fffefa',
    100: '#f2efe8',
    150: '#ebe6dd',
    200: '#e9e5dc',
    250: '#e2ddd4',
    300: '#d3ccc0',
    400: '#bcb4a8',
    500: '#857d72',
  },
  /** Warm inks, light theme, and the light footer band. */
  plum: {
    100: '#ddd6d9',
    300: '#b4abaf',
    400: '#9d9498',
    600: '#5e5659',
    650: '#4a4245',
    700: '#393235',
    900: '#211a1d',
    950: '#1b1618',
  },
  /** Neutral greys, Graphite dark theme. */
  graphite: {
    100: '#e8e8e8',
    300: '#aeaeae',
    400: '#808080',
    500: '#5f5f5f',
    600: '#444444',
    700: '#2c2c2c',
    750: '#252525',
    800: '#232323',
    850: '#1c1c1c',
    900: '#141414',
  },
  violet: {
    25: '#ece7ff',
    50: '#e8e4fa',
    100: '#cbc2f8',
    150: '#c4b9ff',
    200: '#c0b4f5',
    250: '#b69ad0',
    300: '#a99cf0',
    350: '#9a8bea',
    375: '#9a8ae6',
    400: '#8f80e8',
    425: '#8a7ae0',
    450: '#7a6ad2',
    475: '#7262dd',
    500: '#6a58cf',
    525: '#6656d4',
    550: '#5b4bc8',
    600: '#5747b8',
    625: '#6c568c',
    650: '#4d3daa',
    700: '#49399f',
    725: '#40318c',
    750: '#3f3196',
    775: '#3e3290',
    800: '#35277f',
    850: '#33277c',
    900: '#383646',
  },
  red: {
    300: '#d77171',
    400: '#c05050',
    450: '#b84a4a',
    500: '#a33333',
    600: '#892828',
    700: '#7a2b2b',
    800: '#651b1b',
  },
  green: {
    300: '#7fb488',
    600: '#3f7048',
  },
  amber: {
    300: '#ddb66f',
    500: '#9b6a25',
    600: '#8a5b16',
  },
  blue: {
    300: '#70acd0',
    600: '#2f6f8f',
  },
  teal: {
    300: '#83aaa3',
    600: '#55716d',
  },
  rose: {
    300: '#c58e9b',
    600: '#83565d',
  },
  /** Leaderboard podium medals. */
  gold: { 100: '#f3e2a6', 300: '#f2d77c', 800: '#4d3f12', 900: '#5c4300' },
  silver: { 100: '#e3e3ea', 200: '#dcdce6', 800: '#565660', 850: '#3d3d4a' },
  bronze: { 100: '#f0d3bd', 300: '#eebc98', 800: '#4a2c1b', 850: '#6b3412' },
  /** Shadow and scrim bases. */
  shade: {
    plum: 'rgb(33 26 29 / 16%)',
    violet: 'rgb(124 104 224 / 34%)',
    scrimLight: 'rgb(33 26 29 / 28%)',
    scrimDark: 'rgb(0 0 0 / 38%)',
    forcedScrim: 'rgb(0 0 0 / 50%)',
    violetTint: 'rgb(143 128 232 / 14%)',
  },
} as const
