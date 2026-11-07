import * as stylex from '@stylexjs/stylex'
import { Link } from '@tanstack/react-router'
import { Button, buttonStyles, iconSizes } from 'paper-ui'
import { font, size, space, type } from 'paper-ui/tokens.stylex'
import { ChevronLeftIcon, ChevronRightIcon } from 'paper-ui/icons'
import type { ReactNode } from 'react'

const styles = stylex.create({
  stepper: { gap: space[4], alignItems: 'center', display: 'flex', },
  year: {
    fontFamily: font.display,
    fontSize: type.sectionSize,
    fontVariantNumeric: 'tabular-nums',
    lineHeight: type.sectionLineHeight,
    minInlineSize: size[72],
    textAlign: 'center',
  },
})

/**
 * A year with previous and next steps, from the account's first year to the current one. A year outside that range
 * steps back into it.
 */
export function YearStepper({ year, first, current, href, label }: {
  year: number
  first: number
  current: number
  href: (year: number) => string
  /** Names the group, for example `Profile year`. */
  label: string
}) {
  const previous = year > first ? Math.min(year - 1, current) : undefined
  const next = year < current ? Math.max(year + 1, first) : undefined
  const step = (target: number | undefined, name: string, icon: ReactNode) => target === undefined
    ? <Button variant="ghost" aria-label={name} disabled>{icon}</Button>
    : <Link {...stylex.props(...buttonStyles({ variant: 'ghost' }))} aria-label={`${name}, ${target}`} to={href(target)}>{icon}</Link>

  return (
    <div {...stylex.props(styles.stepper)} role="group" aria-label={label}>
      {step(previous, 'Previous year', <ChevronLeftIcon {...stylex.props(iconSizes.default)} aria-hidden="true" />)}
      <strong aria-live="polite" {...stylex.props(styles.year)}>{year}</strong>
      {step(next, 'Next year', <ChevronRightIcon {...stylex.props(iconSizes.default)} aria-hidden="true" />)}
    </div>
  )
}
