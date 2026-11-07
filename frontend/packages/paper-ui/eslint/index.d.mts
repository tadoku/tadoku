import type { Linter, Rule } from 'eslint'

export const paperPlugin: { readonly rules: Readonly<Record<'tokens-only' | 'no-raw-styling', Rule.RuleModule>> }

/** Paper's StyleX lint rules for the given source globs. */
export function paperStyleConfig(files?: readonly string[]): Linter.Config[]
