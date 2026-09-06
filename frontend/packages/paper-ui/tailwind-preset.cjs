/** @type {import('tailwindcss').Config} */
module.exports = Object.freeze({
  // Paper supplies its own reset and component defaults.
  corePlugins: { preflight: false },
  theme: {
    extend: {
      // Extend instead of replace: w-5, p-10 and the rest of Tailwind stay available.
      spacing: {
        1: 'var(--paper-space-1)',
        2: 'var(--paper-space-2)',
        3: 'var(--paper-space-3)',
        4: 'var(--paper-space-4)',
        6: 'var(--paper-space-6)',
        8: 'var(--paper-space-8)',
        12: 'var(--paper-space-12)',
      },
      gap: { inline: 'var(--paper-inline-gap)' },
      colors: {
        canvas: 'var(--paper-color-surface-canvas)',
        paper: 'var(--paper-color-surface-paper)',
        raised: 'var(--paper-color-surface-raised)',
        overlay: 'var(--paper-color-surface-overlay)',
        scrim: 'var(--paper-color-surface-scrim)',
        ink: 'var(--paper-color-text-ink)',
        muted: 'var(--paper-color-text-muted)',
        inverse: 'var(--paper-color-text-inverse)',
        link: 'var(--paper-color-text-link)',
        rule: {
          DEFAULT: 'var(--paper-color-rule-default)',
          subtle: 'var(--paper-color-rule-subtle)',
          strong: 'var(--paper-color-rule-strong)',
        },
        'field-edge': 'var(--paper-color-rule-field-edge)',
        'action-edge': 'var(--paper-color-rule-action-edge)',
        'destructive-edge': 'var(--paper-color-rule-destructive-action-edge)',
        action: {
          DEFAULT: 'var(--paper-color-action-default)',
          hover: 'var(--paper-color-action-hover)',
          active: 'var(--paper-color-action-active)',
          soft: 'var(--paper-color-action-soft)',
          text: 'var(--paper-color-action-text)',
        },
        'neutral-hover': 'var(--paper-color-action-neutral-hover)',
        destructive: {
          DEFAULT: 'var(--paper-color-action-destructive)',
          hover: 'var(--paper-color-action-destructive-hover)',
        },
        information: 'var(--paper-color-status-information)',
        success: 'var(--paper-color-status-success)',
        warning: 'var(--paper-color-status-warning)',
        danger: 'var(--paper-color-status-danger)',
        focus: 'var(--paper-color-focus-ring)',
        'focus-offset': 'var(--paper-color-focus-offset)',
        chart: {
          1: 'var(--paper-color-chart-1)',
          2: 'var(--paper-color-chart-2)',
          3: 'var(--paper-color-chart-3)',
          4: 'var(--paper-color-chart-4)',
          5: 'var(--paper-color-chart-5)',
          6: 'var(--paper-color-chart-6)',
          7: 'var(--paper-color-chart-7)',
          8: 'var(--paper-color-chart-8)',
        },
      },
      fontFamily: {
        sans: ['var(--paper-font-body)'],
        serif: ['var(--paper-font-display)'],
        mono: ['var(--paper-font-monospace)'],
      },
      minHeight: {
        control: 'var(--paper-control-height)',
      },
      transitionDuration: {
        quick: 'var(--paper-motion-quick)',
        standard: 'var(--paper-motion-standard)',
        deliberate: 'var(--paper-motion-deliberate)',
      },
    },
  },
})
