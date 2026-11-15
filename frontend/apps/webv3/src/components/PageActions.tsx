import { useHydrated, useNavigate } from '@tanstack/react-router'
import { buttonClassName, type ActionMenuItem, type ButtonVariant } from 'paper-ui'
import { iconClassName } from 'paper-ui/icons'
import type { ReactNode } from 'react'
import { OptionsMenu } from './OptionsMenu'
import { AppLink } from './router-link'

/** A page action: a link to another page, or a callback such as opening a dialog. */
export type PageAction = {
  readonly id: string
  readonly label: string
  readonly icon?: ReactNode
  readonly variant?: ButtonVariant
  readonly destructive?: boolean
} & ({ readonly href: string } | { readonly onSelect: () => void })

/**
 * A page header's actions, laid out by the header's width alone, so server HTML already has the right layout.
 * At `actions-full` and wider every action is a button; from `actions-split` the primary ones stay beside an
 * ellipsis menu of the rest; narrower still, the menu holds everything. Must sit inside a `PageHeader`, the container
 * it queries. Secondary actions come first, as quiet actions precede the stronger one. Pass `pageActionsHeader` to the
 * header's `classNames` so the actions stay on the title's row, centred on its first line, instead of wrapping under it.
 */
export function PageActions({ label, primary = [], secondary = [] }: {
  label: string
  primary?: readonly PageAction[]
  secondary?: readonly PageAction[]
}) {
  const navigate = useNavigate()
  const hydrated = useHydrated()
  const item = (action: PageAction): ActionMenuItem => ({
    id: action.id,
    label: action.label,
    icon: action.icon,
    destructive: action.destructive,
    onSelect: 'href' in action ? () => void navigate({ to: action.href }) : action.onSelect,
  })
  const button = (action: PageAction, hidden: string) => {
    const className = buttonClassName({ variant: action.variant ?? 'outline', className: hidden })
    const content = (
      <>
        {action.icon ? <span className={iconClassName('compact', '*:size-full')} aria-hidden="true">{action.icon}</span> : null}
        {action.label}
      </>
    )
    return 'href' in action
      ? <AppLink key={action.id} className={className} href={action.href}>{content}</AppLink>
      : <button key={action.id} type="button" className={className} disabled={!hydrated} onClick={action.onSelect}>{content}</button>
  }

  return (
    <span className="flex items-center gap-inline">
      {secondary.map(action => button(action, '@max-actions-full:hidden'))}
      {primary.map(action => button(action, '@max-actions-split:hidden'))}
      {secondary.length ? (
        <span className="hidden @actions-split:@max-actions-full:inline-flex">
          <OptionsMenu label={label} items={secondary.map(item)} />
        </span>
      ) : null}
      <span className="inline-flex @actions-split:hidden">
        <OptionsMenu label={label} items={[...primary, ...secondary].map(item)} />
      </span>
    </span>
  )
}
