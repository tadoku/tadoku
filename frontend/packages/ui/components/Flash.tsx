import { ComponentType, ReactNode } from 'react'
import Link from 'next/link'

interface Props {
  style: 'success' | 'warning' | 'error' | 'info'
  children: ReactNode
  IconComponent?: ComponentType<any>
  href?: string
  onClick?: () => void
  className?: string
}

export const Flash = ({
  style,
  children,
  IconComponent,
  href,
  onClick,
  className,
}: Props) => {
  if (onClick || href) {
    return (
      <Link
        href={href ?? '#'}
        onClick={onClick}
        className={`flash ${style} ${className ?? ''}`}
      >
        {IconComponent ? <IconComponent className="mr-2 h-5" /> : null}
        {children}
      </Link>
    )
  }
  return (
    <div className={`flash ${style} ${className ?? ''}`}>
      {IconComponent ? <IconComponent className="mr-2 h-5" /> : null}
      {children}
    </div>
  )
}
