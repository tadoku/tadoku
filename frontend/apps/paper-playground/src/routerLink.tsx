import * as stylex from '@stylexjs/stylex'
import { linkStyles } from 'paper-ui'
import { Link } from 'react-router-dom'
import type { NavigationLinkProps } from 'paper-ui'

export const renderLink = ({ href, ...props }: NavigationLinkProps) => <Link {...stylex.props(linkStyles.base)} to={href} {...props} />
