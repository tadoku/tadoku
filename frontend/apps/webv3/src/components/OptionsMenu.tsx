import { useHydrated } from '@tanstack/react-router'
import { ActionMenu, buttonClassName, type ActionMenuItem } from 'paper-ui'
import { EllipsisHorizontalIcon, iconClassName } from 'paper-ui/icons'

/** An ellipsis menu of page actions. Server HTML renders its trigger disabled until the menu can open. */
export function OptionsMenu({ label, items }: { label: string; items: ActionMenuItem[] }) {
  const hydrated = useHydrated()
  if (hydrated) return <ActionMenu label={label} items={items} iconOnly />

  return (
    <button
      type="button"
      className={buttonClassName({
        variant: 'outline',
        className: 'w-target min-h-target min-w-target p-0 disabled:cursor-progress disabled:opacity-full',
      })}
      aria-label={label}
      title={label}
      disabled
    >
      <EllipsisHorizontalIcon className={iconClassName()} aria-hidden="true" />
    </button>
  )
}
