import { Link } from 'react-router-dom'
import type { NavigationLinkProps } from 'paper-ui'

export const renderLink = ({ href, ...props }: NavigationLinkProps) => <Link className="text-link" to={href} {...props} />
