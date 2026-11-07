import type { ReactNode } from 'react'

const podium = {
  1: 'bg-podium-1 text-on-podium-1',
  2: 'bg-podium-2 text-on-podium-2',
  3: 'bg-podium-3 text-on-podium-3',
} as const

/** A participant's rank; the top three sit on gold, silver and bronze marks. */
export function Rank({ rank, children }: { rank: number; children?: ReactNode }) {
  const medal = rank >= 1 && rank <= 3 ? podium[rank as 1 | 2 | 3] : null
  return (
    <span
      className={medal
        ? `standings-rank inline-grid h-7 min-w-7 place-items-center rounded-full px-1 font-bold tabular-nums [&_abbr]:no-underline ${medal}`
        : 'standings-rank text-muted [&_abbr]:no-underline'}
      data-podium={medal ? rank : undefined}
    >
      {children ?? rank}
    </span>
  )
}

/** A participant's name, with a filled You pill on the viewer's own row. */
export function Person({ children, you }: { children: ReactNode; you?: boolean }) {
  return (
    <span className="standings-person inline-flex flex-wrap items-baseline gap-x-2 gap-y-1">
      {children}
      {you ? (
        <span className="relative -top-px inline-block rounded-pill bg-action px-2 py-px text-xs leading-150 font-bold tracking-wide text-on-action">
          You
        </span>
      ) : null}
    </span>
  )
}

/** The link on a participant's name: ink, underlined with a rule until hovered. */
export const personLink = 'text-ink decoration-rule wrap-anywhere hover:text-link hover:decoration-current'

/** The viewer's row: an accent-soft band. */
export const currentRow = 'bg-action-soft hover:bg-current-row-hover'
