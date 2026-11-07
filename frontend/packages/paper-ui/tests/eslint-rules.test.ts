import { RuleTester } from 'eslint'
import { paperPlugin } from '../eslint/index.mjs'

RuleTester.describe = describe
RuleTester.it = it
RuleTester.itOnly = it.only

const tester = new RuleTester({ languageOptions: { ecmaVersion: 2022, sourceType: 'module' } })
const styles = (value: string) => `import * as stylex from '@stylexjs/stylex'\nstylex.create({ a: { paddingInline: ${value} } })`

tester.run('paper/tokens-only', paperPlugin.rules['tokens-only'], {
  valid: [
    styles('space[16]'),
    // Base UI sets these on its popups; they are runtime geometry, not design values.
    `import * as stylex from '@stylexjs/stylex'\nstylex.create({ a: { inlineSize: 'var(--anchor-width)', transformOrigin: 'var(--transform-origin)' } })`,
  ],
  invalid: [
    { code: styles("'var(--app-gutter)'"), errors: [{ messageId: 'customProperty' }] },
    { code: styles('`calc(${space[4]} + var(--app-gutter))`'), errors: [{ messageId: 'customProperty' }] },
  ],
})

const jsx = { languageOptions: { ecmaVersion: 2022 as const, sourceType: 'module' as const, parserOptions: { ecmaFeatures: { jsx: true } } } }
const jsxTester = new RuleTester(jsx)

jsxTester.run('paper/no-raw-styling', paperPlugin.rules['no-raw-styling'], {
  valid: ['<Table tableXstyle={styles.table} />', '<div {...stylex.props(styles.a)} />'],
  invalid: [
    { code: '<div className="page" />', errors: [{ messageId: 'className' }] },
    { code: '<Table tableClassName="standings-table" />', errors: [{ messageId: 'className' }] },
    { code: '<div style={{ color: "red" }} />', errors: [{ messageId: 'style' }] },
  ],
})

jsxTester.run('paper/tokens-only in JSX size props', paperPlugin.rules['tokens-only'], {
  valid: ['<Table minWidth={size[512]} />', '<Table minWidth="0" />', '<img width="36" height="36" />'],
  invalid: [{ code: '<Table minWidth="32rem" />', errors: [{ messageId: 'literal' }] }],
})
