import { cn } from 'paper-ui'
import { useLinks } from '../lib/links'

/** The first letters of the first two words, or the first two letters of a one-word name. */
const initials = (name: string) => {
  const words = name.trim().split(/\s+/)
  return (words.length > 1 ? words[0][0] + words[1][0] : name.trim().slice(0, 2)).toLocaleUpperCase()
}

const sizes = {
  sm: 'size-5 text-2xs text-muted',
  md: 'size-6 text-2xs text-muted',
  lg: 'size-16 text-xl tracking-wide',
} as const

/**
 * A learner's Gravatar, which the API proxies so their email never reaches the browser, over their initials.
 * The initials show while the image loads and stay when the learner has no Gravatar or it fails. Decorative: the name is always beside it.
 */
export function Avatar({ userId, name, size, className }: {
  userId: string
  name: string
  size: keyof typeof sizes
  className?: string
}) {
  const links = useLinks()

  return (
    <span
      className={cn(
        'relative inline-grid flex-none place-items-center overflow-hidden border border-rule bg-paper font-body font-semibold leading-none',
        sizes[size],
        className,
      )}
      aria-hidden="true"
    >
      {size === 'sm' ? initials(name).slice(0, 1) : initials(name)}
      {/* A background, unlike an img, draws nothing when it fails, even before hydration. */}
      <span className="absolute inset-0 bg-cover bg-center" style={{ backgroundImage: `url("${links.avatar(userId)}")` }} />
    </span>
  )
}
