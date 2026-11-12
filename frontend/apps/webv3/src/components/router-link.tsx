import { Link } from '@tanstack/react-router'
import type { NavigationLinkProps } from 'paper-ui'
import { parseSearch } from '../lib/search'

/**
 * Renders Paper navigation links: same-host paths, with any query and fragment, through the router; other origins as
 * plain links.
 */
export function renderLink({ href, ...props }: NavigationLinkProps) {
  if (!href.startsWith('/') || href.startsWith('//')) return <a href={href} {...props} />

  const url = new URL(href, 'http://same-host')
  return <Link to={url.pathname} search={parseSearch(url.search)} hash={url.hash.slice(1) || undefined} {...props} />
}

/** A link written in JSX, rendered as `renderLink` renders it. */
export const AppLink = renderLink
