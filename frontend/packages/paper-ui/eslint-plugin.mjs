/**
 * Keeps Paper's styling on its tokens. `paper/no-arbitrary-values` rejects Tailwind arbitrary values such as
 * `p-[13px]`, `bg-[#fff]` or `[mask:url(x)]` in class names. The only arbitrary values it allows read a custom
 * property, as `w-(--anchor-width)` or `[background:var(--effect-table-scroll)]`, and that property must be a Paper
 * token or one Base UI sets at runtime. Arbitrary variants such as `[&_a]:` select elements, not values, so they
 * stay available.
 */
const tokenPrefixes = [
  '--color-',
  '--spacing-',
  '--container-',
  '--height-',
  '--text-',
  '--leading-',
  '--effect-',
]
const runtimeProperties = new Set(['--anchor-width', '--transform-origin', '--available-height'])
const classCallees = new Set(['cn', 'tv', 'clsx'])

const allowedProperty = (name) =>
  runtimeProperties.has(name) || tokenPrefixes.some((prefix) => name.startsWith(prefix))

/** Splits `md:[&_a]:p-2` into its variants and utility, ignoring colons inside brackets and parentheses. */
function utilityOf(candidate) {
  let depth = 0
  let start = 0
  for (let index = 0; index < candidate.length; index += 1) {
    const character = candidate[index]
    if (character === '[' || character === '(') depth += 1
    else if (character === ']' || character === ')') depth -= 1
    else if (character === ':' && depth === 0) start = index + 1
  }
  return candidate.slice(start).replace(/^!|!$/g, '')
}

export function arbitraryValueProblem(candidate) {
  const utility = utilityOf(candidate)
  if (!/[[(]/.test(utility)) return null
  const reference =
    utility.match(/^-?[a-z][a-z0-9-]*-\((--[\w-]+)\)$/)?.[1] ??
    utility.match(/^-?[a-z][a-z0-9-]*-\[var\((--[\w-]+)\)\]$/)?.[1] ??
    utility.match(/^\[[a-z-]+:var\((--[\w-]+)\)\]$/)?.[1]
  if (reference && allowedProperty(reference)) return null
  return reference
    ? `${candidate} reads ${reference}, which is neither a Paper token nor a Base UI runtime property`
    : `${candidate} is an arbitrary value; add a token in paper-ui/src/tokens instead`
}

function isClassContext(node) {
  for (let current = node.parent; current; current = current.parent) {
    if (current.type === 'JSXAttribute') return ['class', 'className', 'classNames'].includes(current.name.name)
    if (current.type === 'CallExpression') {
      const callee = current.callee.type === 'Identifier' ? current.callee.name : current.callee.property?.name
      if (classCallees.has(callee) || /ClassName$/.test(callee ?? '')) return true
    }
    if (current.type === 'VariableDeclarator') return /[cC]lass(Name|es)?$/.test(current.id.name ?? '')
    if (current.type === 'Property' && /[cC]lass(Name)?$/.test(current.key?.name ?? '')) return true
    if (current.type === 'Program' || current.type === 'FunctionDeclaration') return false
  }
  return false
}

const noArbitraryValues = {
  meta: {
    type: 'problem',
    docs: { description: 'Disallow Tailwind arbitrary values that bypass Paper tokens' },
    schema: [],
  },
  create(context) {
    const inspect = (node, text) => {
      if (!isClassContext(node)) return
      for (const candidate of text.split(/\s+/)) {
        if (!candidate) continue
        const problem = arbitraryValueProblem(candidate)
        if (problem) context.report({ node, message: problem })
      }
    }
    return {
      Literal(node) {
        if (typeof node.value === 'string') inspect(node, node.value)
      },
      TemplateElement(node) {
        inspect(node, node.value.cooked ?? '')
      },
    }
  },
}

/** Rejects literal values in a JSX `style` object; computed values such as a measured width stay allowed. */
const noLiteralStyles = {
  meta: {
    type: 'problem',
    docs: { description: 'Disallow literal values in inline styles' },
    schema: [],
  },
  create(context) {
    return {
      JSXAttribute(node) {
        if (node.name.name !== 'style' || node.value?.type !== 'JSXExpressionContainer') return
        const visit = (expression) => {
          if (expression?.type === 'ConditionalExpression') {
            visit(expression.consequent)
            visit(expression.alternate)
          } else if (expression?.type === 'ObjectExpression') {
            for (const property of expression.properties) {
              if (property.type === 'Property' && property.value.type === 'Literal') {
                context.report({ node: property, message: 'Inline style literals bypass Paper tokens; use a utility class' })
              }
            }
          }
        }
        visit(node.value.expression)
      },
    }
  },
}

export default {
  meta: { name: 'paper' },
  rules: { 'no-arbitrary-values': noArbitraryValues, 'no-literal-styles': noLiteralStyles },
}
