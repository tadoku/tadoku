import { cn } from 'paper-ui'
import { useState } from 'react'
import { useLinks } from '../lib/links'

/** The first letters of the first two words, or the first two letters of a one-word name. */
const initials = (name: string) => {
  const words = name.trim().split(/\s+/)
  return (words.length > 1 ? words[0][0] + words[1][0] : name.trim().slice(0, 2)).toUpperCase()
}

/** Each size's rendered width and height in CSS pixels, matching its `size-*` class, for the image's dimensions. */
const pixels = { sm: 20, md: 24, nav: 32, lg: 64 } as const

const sizes = {
  /** Inline in a line of text: centred between its lowercase and capital letters. */
  sm: '-top-px size-5 align-middle text-2xs text-muted',
  md: 'size-6 text-2xs text-muted',
  nav: 'size-8 text-xs text-muted',
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
        'relative inline-grid flex-none place-items-center overflow-hidden rounded-pill border border-rule bg-paper font-body font-semibold leading-none',
        sizes[size],
        className,
      )}
      aria-hidden="true"
    >
      {size === 'sm' ? initials(name).slice(0, 1) : initials(name)}
      <AvatarImage src={links.avatar(userId)} pixels={pixels[size]} />
    </span>
  )
}

/**
 * The Gravatar over the initials, loaded lazily. It stays invisible until it has loaded, so a learner without one, or a
 * failed load, shows the initials rather than a broken image; without JavaScript the initials always show.
 */
function AvatarImage({ src, pixels }: { src: string; pixels: number }) {
  const [loaded, setLoaded] = useState(false)
  // An image that finished before hydration fired its load event before React listened.
  const seen = (image: HTMLImageElement | null) => {
    if (image?.complete && image.naturalWidth > 0) setLoaded(true)
  }

  return (
    <img
      ref={seen}
      className={cn('absolute inset-0 size-full object-cover', !loaded && 'invisible')}
      src={src}
      alt=""
      width={pixels}
      height={pixels}
      loading="lazy"
      decoding="async"
      onLoad={() => setLoaded(true)}
    />
  )
}
