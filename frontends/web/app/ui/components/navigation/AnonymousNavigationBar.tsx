import React from 'react'
import Link from 'next/link'
import { NavigationBarLink } from './index'

export const AnonymousNavigationBar = () => (
  <>
    <Link href="/sign-in">
      <NavigationBarLink href="">Sign in</NavigationBarLink>
    </Link>
    <Link href="/register">
      <NavigationBarLink href="">Register</NavigationBarLink>
    </Link>
  </>
)
