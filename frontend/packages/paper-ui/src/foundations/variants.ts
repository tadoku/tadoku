import { createTV, extendTailwindMerge } from 'tailwind-variants/config'
import { theme, type ThemeNamespace } from '../tokens/theme'

const keys = (namespace: ThemeNamespace) => Object.keys(theme[namespace]).map((key) => key.replace('_', '.'))

/** Teaches tailwind-merge Paper's token names, so `text-muted` and `text-label` don't override each other. */
export const twMergeConfig = {
  override: {
    theme: {
      color: [...keys('color'), 'transparent', 'current', 'inherit'],
      font: keys('font'),
      text: keys('text'),
      'font-weight': keys('font-weight'),
      tracking: keys('tracking'),
      leading: keys('leading'),
      breakpoint: keys('breakpoint'),
      container: keys('container'),
      spacing: keys('spacing'),
      radius: keys('radius'),
      shadow: keys('shadow'),
      blur: keys('blur'),
      animate: keys('animate'),
    },
  },
  extend: {
    classGroups: {
      z: [{ z: keys('z-index') }],
      opacity: [{ opacity: keys('opacity') }],
      scale: [{ scale: keys('scale') }],
      duration: [{ duration: keys('transition-duration') }],
      'grid-cols': [{ 'grid-cols': keys('grid-template-columns') }],
      h: [{ h: keys('height') }],
      'max-h': [{ 'max-h': keys('height') }],
      'min-h': [{ 'min-h': keys('height') }],
      'border-w': [{ border: keys('border-width') }],
      'border-w-x': [{ 'border-x': keys('border-width') }],
      'border-w-y': [{ 'border-y': keys('border-width') }],
      'border-w-s': [{ 'border-s': keys('border-width') }],
      'border-w-e': [{ 'border-e': keys('border-width') }],
      'border-w-t': [{ 'border-t': keys('border-width') }],
      'border-w-r': [{ 'border-r': keys('border-width') }],
      'border-w-b': [{ 'border-b': keys('border-width') }],
      'border-w-l': [{ 'border-l': keys('border-width') }],
      'outline-w': [{ outline: keys('outline-width') }],
      'outline-offset': [{ 'outline-offset': keys('outline-offset') }],
    },
  },
}

/** Merges class names, letting a caller's `className` override a component's own utilities. */
export const cn = extendTailwindMerge(twMergeConfig)

/** Paper's variant helper: every component recipe is a `tv()` call, so callers' classes merge predictably. */
export const tv = createTV({ twMerge: cn })

export type { VariantProps } from 'tailwind-variants'
