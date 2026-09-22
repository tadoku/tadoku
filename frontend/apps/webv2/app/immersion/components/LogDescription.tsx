import Link from 'next/link'
import { useId, useState } from 'react'

interface Props {
  description: string
  tags: string[]
  href?: string
}

export function LogDescription({ description, tags, href }: Props) {
  const descriptionId = useId()
  const [revealed, setRevealed] = useState(false)
  const [hovered, setHovered] = useState(false)

  if (!tags.some(tag => tag.toLowerCase() === 'nsfw')) {
    return href ? (
      <Link className="reset" href={href}>
        {description}
      </Link>
    ) : (
      <span>{description}</span>
    )
  }

  const visible = revealed || hovered

  return (
    <button
      type="button"
      className="btn ghost relative !h-auto !text-left !font-normal"
      aria-label={
        revealed ? 'Hide NSFW description' : 'Reveal NSFW description'
      }
      aria-expanded={visible}
      aria-describedby={visible ? descriptionId : undefined}
      onPointerEnter={event => {
        if (event.pointerType === 'mouse') setHovered(true)
      }}
      onPointerLeave={() => setHovered(false)}
      onClick={() => setRevealed(!revealed)}
    >
      <span
        id={descriptionId}
        aria-hidden={!visible}
        className={`pointer-events-none ${visible ? '' : 'blur-md select-none'}`}
      >
        {description}
      </span>
      {!visible && (
        <span
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 flex items-center justify-center text-xs font-bold"
        >
          NSFW
        </span>
      )}
    </button>
  )
}
