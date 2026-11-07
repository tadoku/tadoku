import * as stylex from '@stylexjs/stylex'

type StyleArgument = stylex.StyleXArray<
  stylex.CompiledStyles | null | undefined | boolean | Readonly<[stylex.CompiledStyles, stylex.InlineStyles]>
>

/** A component's `xstyle` prop: compiled styles, or nested arrays of them with falsy entries skipped. */
export type StyleXProp = StyleArgument

/** `stylex.props` plus a caller's className, for components that accept both. */
export function sx(className: string | undefined, ...styles: readonly StyleArgument[]) {
  const props = stylex.props(...styles)
  if (!className) return props
  return { ...props, className: props.className ? `${props.className} ${className}` : className }
}
