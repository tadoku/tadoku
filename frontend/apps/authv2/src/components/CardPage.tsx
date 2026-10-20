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
    <div className="card-page">
      <Surface className="card-page__card">
        <PageHeader title={title} lead={lead} className="card-page__header" />
        {children}
      </Surface>
      {footer ? <p className="card-page__footer">{footer}</p> : null}
    </div>
  )
}
