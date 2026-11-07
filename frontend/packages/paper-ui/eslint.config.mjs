import eslint from '@eslint/js'
import tseslint from 'typescript-eslint'
import { paperStyleConfig } from './eslint/index.mjs'

export default tseslint.config(
  { ignores: ['dist/**', 'coverage/**'] },
  ...paperStyleConfig(['src/**/*.{ts,tsx}']),
  // Tests pass plain class names to prove components forward them.
  { files: ['**/*.test.{ts,tsx}'], rules: { 'paper/no-raw-styling': 'off' } },
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
    files: ['scripts/**/*.mjs'],
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
