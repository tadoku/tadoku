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
      className={`standings-rank inline-flex h-7 min-w-7 items-center justify-center rounded-chip px-1 tabular-nums whitespace-nowrap [&_abbr]:no-underline ${medal ? `font-bold ${medal}` : 'text-muted'}`}
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
        <span className="relative -top-px inline-block rounded-chip border border-ink px-1.5 py-px text-xs leading-125 font-semibold text-ink">
          You
        </span>
      ) : null}
    </span>
  )
}

/** The link on a participant's name: row context marks it, so the underline waits for hover. */
export const personLink = 'text-ink no-underline wrap-break-word hover:text-link hover:underline'

/** The viewer's row: an ink rail, distinct from the row hover and clear of the accent. */
export const currentRow = 'font-semibold shadow-viewer-start hover:bg-neutral-hover'
