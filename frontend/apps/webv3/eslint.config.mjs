import eslint from '@eslint/js'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import paper from 'paper-ui/eslint-plugin'

export default tseslint.config(
  { ignores: ['dist/**', '.output/**', '.tanstack/**', 'src/routeTree.gen.ts'] },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },
  {
    files: ['src/**/*.{ts,tsx}', 'server/**/*.ts'],
    ignores: ['src/lib/zod.ts'],
    rules: {
      'no-restricted-imports': ['error', {
        paths: [{ name: 'zod', message: "Import z from src/lib/zod.ts, which turns off Zod's eval probe before any schema is built." }],
        patterns: [{ group: ['zod/*'], message: 'Import z from src/lib/zod.ts.' }],
      }],
    },
  },
  {
    files: ['src/routes/**/*.tsx'],
    rules: { 'react-refresh/only-export-components': 'off' },
  },
  {
    files: ['src/**/*.{ts,tsx}'],
    plugins: { paper },
    rules: { 'paper/no-arbitrary-values': 'error', 'paper/no-literal-styles': 'error' },
  },
)
