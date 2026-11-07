import * as stylex from '@stylexjs/stylex'
import { cloneElement, isValidElement, type ReactElement, type ReactNode } from 'react'
import { icon as iconStyles } from './recipes'
import { sx, type StyleXProp } from './props'

const styles = stylex.create({
  hidden: { visibility: 'hidden' },
})

/** Sizes an icon element to fill its box; other nodes render unchanged. */
export function fillIcon(icon: ReactNode) {
  if (!isValidElement(icon)) return icon
  const element = icon as ReactElement<{ className?: string }>
  return cloneElement(element, { className: sx(element.props.className, iconStyles.fill).className })
}

/** A 20px box, hidden from assistive technology, that centres the icon it holds. */
export function IconSlot({ children, xstyle, hidden }: { children: ReactNode; xstyle?: StyleXProp; hidden?: boolean }) {
  return (
    <span aria-hidden="true" {...stylex.props(iconStyles.slot, hidden && styles.hidden, xstyle)}>
      {fillIcon(children)}
    </span>
  )
}
