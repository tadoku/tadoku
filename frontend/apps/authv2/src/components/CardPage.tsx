import * as stylex from '@stylexjs/stylex'
import { PageHeader, Surface } from 'paper-ui'
import type { ReactNode } from 'react'
import { cardStyles } from './styles'

/** A single focused task, such as logging in, on a narrow card. */
export function CardPage({ title, lead, footer, children }: {
  readonly title: string
  readonly lead?: ReactNode
  readonly footer?: ReactNode
  readonly children: ReactNode
}) {
  return (
    <div {...stylex.props(cardStyles.page)}>
      <Surface inset="none" xstyle={cardStyles.card}>
        <PageHeader title={title} lead={lead} xstyle={cardStyles.header} />
        {children}
      </Surface>
      {footer ? <p {...stylex.props(cardStyles.footer)}>{footer}</p> : null}
    </div>
  )
}
