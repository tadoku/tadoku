import { PageHeader, Surface } from 'paper-ui'
import type { ReactNode } from 'react'

/** A single focused task, such as logging in, on a narrow card. */
export function CardPage({ title, lead, footer, children }: {
  readonly title: string
  readonly lead?: ReactNode
  readonly footer?: ReactNode
  readonly children: ReactNode
}) {
  return (
    <div className="mx-auto w-card">
      <Surface className="p-card-pad max-sm:px-5 max-sm:py-6">
        <PageHeader title={title} lead={lead} className="mb-7" />
        {children}
      </Surface>
      {footer ? <p className="mb-0 mt-6 text-center text-ui text-muted">{footer}</p> : null}
    </div>
  )
}
