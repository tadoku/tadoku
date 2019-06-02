import React from 'react'
import Link from 'next/link'
import { NavigationBarLink } from './index'
import { SignIn } from '../../../session/navigation/SignIn'

export const AnonymousNavigationBar = () => (
  <>
    <SignIn />
    <Link href="/register">
      <NavigationBarLink href="">Register</NavigationBarLink>
    </Link>
  </>
)
