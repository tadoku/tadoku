// ESLint rules that keep Paper a design system: StyleX's own rules, token-only values and no ad-hoc styling.
import stylexPlugin from '@stylexjs/eslint-plugin'

/** Properties whose values must come from `paper-ui/tokens.stylex`, apart from the keywords below. */
const tokenProperty =
  /^(color|fill|stroke|accentColor|caretColor|outlineColor|textDecorationColor|background|backgroundColor|backgroundImage|boxShadow|textShadow|outline|outlineOffset|outlineWidth|border.*(Color|Width|Radius)|borderWidth|borderRadius|margin.*|padding.*|gap|rowGap|columnGap|inset.*|top|right|bottom|left|(min|max)?(Width|Height|InlineSize|BlockSize)|width|height|inlineSize|blockSize|flexBasis|gridTemplateColumns|gridTemplateRows|gridAutoRows|gridAutoColumns|fontSize|fontFamily|fontWeight|lineHeight|letterSpacing|textIndent|textUnderlineOffset|textDecorationThickness|scrollMargin.*|scrollPadding.*|zIndex|opacity|transitionDuration|animationDuration|transform|translate|strokeWidth)$/u

const keywords = new Set([
  '0', 'auto', 'none', 'inherit', 'initial', 'unset', 'currentColor', 'currentcolor', 'transparent', '100%', '50%',
  '-50%', 'fit-content', 'min-content', 'max-content', 'normal', 'solid', 'Canvas', 'CanvasText', 'Highlight',
  'HighlightText', 'GrayText', 'Mark', 'LinkText', 'ButtonText', 'ButtonFace', '100vw', '100vh', '100dvh', '1fr',
  'minmax(0, 1fr)', 'bold', 'stretch',
])

// A length, time or colour written by hand. Unitless numbers, fr tracks, angles, full-viewport lengths and
// 0, 50 and 100 percent stay allowed: they describe geometry, not design decisions.
const literalValue =
  /(^|[^\w$.-])-?(\d*\.)?\d+(px|rem|em|ch|ex|vw|vh|svh|dvh|ms|s)\b|#[\da-f]{3,8}\b|\b(rgba?|hsla?|oklch|oklab)\(|(^|[^\d.])(?!(0|50|100)%)\d+(\.\d+)?%/iu
const fullViewport = /(^|[^\d.])100(vw|vh|svh|dvh)\b/gu
const isLiteralValue = value => literalValue.test(value.replace(fullViewport, '$1full'))

/** Properties that take a bare number, which must also come from a token unless it is 0. */
const numericProperty = /^(opacity|zIndex|fontWeight|lineHeight|flex|flexGrow|flexShrink|order)$/u

// A hand-written custom property bypasses the typed tokens. Base UI sets these on its popups at runtime.
const runtimeProperty = /var\(--(?!(anchor-width|anchor-height|available-width|available-height|transform-origin)\))/u

function checkCustomProperties(context, node, property) {
  if (!node) return
  if (node.type === 'Literal' && typeof node.value === 'string' && runtimeProperty.test(node.value)) {
    context.report({ node, messageId: 'customProperty', data: { property, value: node.value } })
  } else if (node.type === 'TemplateLiteral') {
    for (const quasi of node.quasis) {
      if (runtimeProperty.test(quasi.value.cooked)) {
        context.report({ node: quasi, messageId: 'customProperty', data: { property, value: quasi.value.cooked } })
      }
    }
  } else if (node.type === 'ObjectExpression') {
    for (const condition of node.properties) {
      if (condition.type === 'Property') checkCustomProperties(context, condition.value, property)
    }
  } else if (node.type === 'ConditionalExpression') {
    checkCustomProperties(context, node.consequent, property)
    checkCustomProperties(context, node.alternate, property)
  }
}

const isStylexCall = (node, name) =>
  node.type === 'CallExpression' &&
  node.callee.type === 'MemberExpression' &&
  node.callee.object.name === 'stylex' &&
  node.callee.property.name === name

function checkValue(context, node, property) {
  if (!node) return
  switch (node.type) {
    case 'Literal':
      if (node.value === null) return
      if (typeof node.value === 'number' ? node.value === 0 : keywords.has(String(node.value).trim())) return
      if (
        typeof node.value === 'string' &&
        !isLiteralValue(node.value) &&
        !(numericProperty.test(property) && /^-?\d*\.?\d+$/u.test(node.value.trim()))
      ) {
        return
      }
      context.report({ node, messageId: 'literal', data: { property, value: String(node.value) } })
      return
    case 'TemplateLiteral':
      for (const quasi of node.quasis) {
        if (isLiteralValue(quasi.value.cooked)) {
          context.report({ node: quasi, messageId: 'literal', data: { property, value: quasi.value.cooked } })
        }
      }
      return
    case 'ObjectExpression':
      for (const condition of node.properties) if (condition.type === 'Property') checkValue(context, condition.value, property)
      return
    case 'ConditionalExpression':
      checkValue(context, node.consequent, property)
      checkValue(context, node.alternate, property)
      return
    case 'CallExpression':
      // stylex.firstThatWorks(...) and similar helpers: check their arguments.
      for (const argument of node.arguments) checkValue(context, argument, property)
      return
    default:
  }
}

function checkStyles(context, node) {
  if (!node || node.type !== 'ObjectExpression') return
  for (const entry of node.properties) {
    if (entry.type !== 'Property') continue
    const name = entry.key.type === 'Identifier' ? entry.key.name : String(entry.key.value)
    if (/^[:@[]/u.test(name)) {
      checkStyles(context, entry.value)
      continue
    }
    checkCustomProperties(context, entry.value, name)
    if (tokenProperty.test(name)) checkValue(context, entry.value, name)
  }
}

const tokensOnly = {
  meta: {
    type: 'problem',
    docs: { description: 'Colour, spacing, size, radius, type, layer and motion values come from Paper tokens.' },
    messages: {
      literal: '{{property}} uses the literal "{{value}}". Use a token from paper-ui/tokens.stylex instead.',
      customProperty: '{{property}} reads a custom property in "{{value}}". Use a token from paper-ui/tokens.stylex instead.',
    },
    schema: [],
  },
  create(context) {
    return {
      // Paper components take a few lengths as props, such as a table's minimum width.
      JSXAttribute(node) {
        const name = node.name.name
        const element = node.parent.name
        if (element.type !== 'JSXIdentifier' || !/^[A-Z]/u.test(element.name)) return
        if (typeof name !== 'string' || !/^(min|max)?(Width|Height|width|height)$/u.test(name) || !node.value) return
        checkValue(context, node.value.type === 'JSXExpressionContainer' ? node.value.expression : node.value, name)
      },
      CallExpression(node) {
        if (!isStylexCall(node, 'create') || node.arguments[0]?.type !== 'ObjectExpression') return
        for (const namespace of node.arguments[0].properties) {
          if (namespace.type !== 'Property') continue
          const styles = namespace.value
          checkStyles(context, styles.type === 'ArrowFunctionExpression' ? unwrap(styles.body) : styles)
        }
      },
    }
  },
}

const unwrap = node => (node.type === 'BlockStatement' ? node.body.find(s => s.type === 'ReturnStatement')?.argument : node)

const isStringish = node =>
  node.type === 'Literal' ||
  node.type === 'TemplateLiteral' ||
  (node.type === 'BinaryExpression' && node.operator === '+') ||
  (node.type === 'ConditionalExpression' && (isStringish(node.consequent) || isStringish(node.alternate))) ||
  (node.type === 'LogicalExpression' && isStringish(node.right)) ||
  (node.type === 'CallExpression' && node.callee.type === 'MemberExpression' && node.callee.property.name === 'join')

const noRawStyling = {
  meta: {
    type: 'problem',
    docs: { description: 'Style elements with StyleX: no class name strings and no inline style props.' },
    messages: {
      className: 'Style with {...stylex.props(...)} instead of a className string.',
      style: 'Use a StyleX style, or a dynamic StyleX style for a value that varies at runtime, instead of style.',
    },
    schema: [],
  },
  create(context) {
    return {
      JSXAttribute(node) {
        if (node.name.name === 'style') context.report({ node, messageId: 'style' })
        if (typeof node.name.name !== 'string' || !/(^c|C)lassName$/u.test(node.name.name) || !node.value) return
        const value = node.value.type === 'JSXExpressionContainer' ? node.value.expression : node.value
        if (isStringish(value)) context.report({ node, messageId: 'className' })
      },
    }
  },
}

export const paperPlugin = { rules: { 'tokens-only': tokensOnly, 'no-raw-styling': noRawStyling } }

/** Flat config entries for the given source globs. */
export function paperStyleConfig(files = ['**/*.{ts,tsx}']) {
  return [
    {
      files,
      plugins: { '@stylexjs': stylexPlugin, paper: paperPlugin },
      rules: {
        '@stylexjs/valid-styles': 'error',
        '@stylexjs/enforce-extension': 'error',
        '@stylexjs/no-unused': 'error',
        '@stylexjs/sort-keys': 'error',
        '@stylexjs/valid-shorthands': 'error',
        '@stylexjs/no-legacy-contextual-styles': 'error',
        'paper/tokens-only': 'error',
        'paper/no-raw-styling': 'error',
      },
    },
  ]
}
