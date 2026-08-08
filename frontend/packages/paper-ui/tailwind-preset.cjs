/** @type {import('tailwindcss').Config} */
module.exports = Object.freeze({
  theme: {
    extend: {
      colors: {
        canvas: 'var(--paper-color-surface-canvas)',
        paper: 'var(--paper-color-surface-paper)',
        ink: 'var(--paper-color-text-ink)',
        muted: 'var(--paper-color-text-muted)',
        action: 'var(--paper-color-action-default)',
        danger: 'var(--paper-color-status-danger)',
      },
      fontFamily: {
        sans: ['Open Sans', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        serif: ['Merriweather', 'ui-serif', 'Georgia', 'serif'],
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
