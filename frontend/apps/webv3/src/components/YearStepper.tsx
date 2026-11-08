import { Link } from '@tanstack/react-router'
import { Button, buttonClassName } from 'paper-ui'
import { ChevronLeftIcon, ChevronRightIcon, iconClassName } from 'paper-ui/icons'
import type { ReactNode } from 'react'

/**
 * A year with previous and next steps, from the account's first year to the current one. A year outside that range
 * steps back into it.
 */
export function YearStepper({ year, first, current, href, label, compact = false }: {
  year: number
  first: number
  current: number
  href: (year: number) => string
  /** Names the group, for example `Profile year`. */
  label: string
  /** A quiet footer control that names the neighbouring years in its steps. */
  compact?: boolean
}) {
  const previous = year > first ? Math.min(year - 1, current) : undefined
  const next = year < current ? Math.max(year + 1, first) : undefined
  const step = (target: number | undefined, name: string, icon: ReactNode, before: boolean) => target === undefined
    ? <Button variant="ghost" aria-label={name} disabled>{icon}</Button>
    : (
      <Link className={buttonClassName({ variant: 'ghost' })} aria-label={`${name}, ${target}`} to={href(target)}>
        {compact && !before ? target : null}{icon}{compact && before ? target : null}
      </Link>
    )

  return (
    <div className="flex items-center gap-1" role="group" aria-label={label}>
      {step(previous, 'Previous year', <ChevronLeftIcon className={iconClassName()} aria-hidden="true" />, true)}
      <strong
        aria-live="polite"
        className={compact ? 'px-1 text-center font-semibold tabular-nums' : 'min-w-18 text-center font-display text-section leading-section tabular-nums'}
      >
        {year}
      </strong>
      {step(next, 'Next year', <ChevronRightIcon className={iconClassName()} aria-hidden="true" />, false)}
    </div>
  )
}
