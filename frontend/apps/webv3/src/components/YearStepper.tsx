import { Link } from '@tanstack/react-router'
import { Button, buttonClassName } from 'paper-ui'
import { ChevronLeftIcon, ChevronRightIcon } from 'paper-ui/icons'
import type { ReactNode } from 'react'

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
    : <Link className={buttonClassName({ variant: 'ghost' })} aria-label={`${name}, ${target}`} to={href(target)}>{icon}</Link>

  return (
    <div className="year-stepper" role="group" aria-label={label}>
      {step(previous, 'Previous year', <ChevronLeftIcon className="paper-icon-default" aria-hidden="true" />)}
      <strong aria-live="polite">{year}</strong>
      {step(next, 'Next year', <ChevronRightIcon className="paper-icon-default" aria-hidden="true" />)}
    </div>
  )
}
