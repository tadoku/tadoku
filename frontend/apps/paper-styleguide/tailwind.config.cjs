const paper = require('paper-ui/tailwind-preset')

/** @type {import('tailwindcss').Config} */
module.exports = {
  presets: [paper],
  content: ['./index.html', './src/**/*.{ts,tsx}', '../../packages/paper-ui/src/catalog/examples/**/*.tsx'],
}
