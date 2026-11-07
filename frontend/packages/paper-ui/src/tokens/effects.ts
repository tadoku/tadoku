/**
 * Composite values that no single Tailwind utility can express. Use them as `[background:var(--effect-name)]`;
 * the ESLint rule allows an arbitrary value only when it reads a token like this.
 */
export const effects = {
  /** Edge shades on a sideways-scrolling table on the paper surface. */
  'table-scroll':
    'linear-gradient(to right, var(--color-paper) 40%, transparent) left / 2rem 100% no-repeat local, ' +
    'linear-gradient(to left, var(--color-paper) 40%, transparent) right / 2rem 100% no-repeat local, ' +
    'radial-gradient(farthest-side at 0 50%, var(--color-table-shade), transparent) left / 0.75rem 100% no-repeat scroll, ' +
    'radial-gradient(farthest-side at 100% 50%, var(--color-table-shade), transparent) right / 0.75rem 100% no-repeat scroll',
  /** The same shades for a table on the canvas, such as one in prose. */
  'canvas-table-scroll':
    'linear-gradient(to right, var(--color-canvas) 40%, transparent) left / 2rem 100% no-repeat local, ' +
    'linear-gradient(to left, var(--color-canvas) 40%, transparent) right / 2rem 100% no-repeat local, ' +
    'radial-gradient(farthest-side at 0 50%, var(--color-canvas-table-shade), transparent) left / 0.75rem 100% no-repeat scroll, ' +
    'radial-gradient(farthest-side at 100% 50%, var(--color-canvas-table-shade), transparent) right / 0.75rem 100% no-repeat scroll',
} as const
