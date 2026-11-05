import { Link } from '@tanstack/react-router'
import type { NavigationLinkProps } from 'paper-ui'

/** Renders Paper navigation links: same-host paths through the router, other origins as plain links. */
export function renderLink({ href, ...props }: NavigationLinkProps) {
  return href.startsWith('/') ? <Link to={href} {...props} /> : <a href={href} {...props} />
}
