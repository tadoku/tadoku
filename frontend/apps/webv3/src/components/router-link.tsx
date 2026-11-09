import { Link } from '@tanstack/react-router'
import type { NavigationLinkProps } from 'paper-ui'

/**
 * Renders Paper navigation links: same-host paths, with any query and fragment, through the router; other origins as
 * plain links.
 */
export function renderLink({ href, ...props }: NavigationLinkProps) {
  if (!href.startsWith('/') || href.startsWith('//')) return <a href={href} {...props} />

  const url = new URL(href, 'http://same-host')
  return <Link to={url.pathname} search={Object.fromEntries(url.searchParams)} hash={url.hash.slice(1) || undefined} {...props} />
}

/** A link written in JSX, rendered as `renderLink` renders it. */
export const AppLink = renderLink
