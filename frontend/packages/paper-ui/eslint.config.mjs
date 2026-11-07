import eslint from '@eslint/js'
import tseslint from 'typescript-eslint'
import paper from './eslint-plugin.mjs'

export default tseslint.config(
  { ignores: ['dist/**', 'coverage/**'] },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parserOptions: {
        ecmaFeatures: { jsx: true },
      },
    },
  },
  {
    files: ['src/**/*.{ts,tsx}'],
    plugins: { paper },
    rules: { 'paper/no-arbitrary-values': 'error', 'paper/no-literal-styles': 'error' },
  },
  {
    // The document base styles elements Paper does not render, so it needs arbitrary properties.
    files: ['src/foundations/document.ts'],
    rules: { 'paper/no-arbitrary-values': 'off' },
  },
  {
    files: ['scripts/**/*.mjs', '*.mjs'],
    languageOptions: {
      globals: {
        URL: 'readonly',
        console: 'readonly',
      },
    },
  },
  {
    files: ['*.cjs'],
    languageOptions: {
      globals: {
        module: 'readonly',
      },
    },
  },
)
