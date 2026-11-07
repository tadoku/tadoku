import eslint from '@eslint/js'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import paper from 'paper-ui/eslint-plugin'

export default tseslint.config(
  { ignores: ['dist/**', 'coverage/**'] },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['*.cjs'],
    languageOptions: { globals: { module: 'readonly', require: 'readonly' } },
    rules: { '@typescript-eslint/no-require-imports': 'off' },
  },
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
    files: ['src/**/*.{ts,tsx}'],
    plugins: { paper },
    rules: { 'paper/no-arbitrary-values': 'error', 'paper/no-literal-styles': 'error' },
  },
)
