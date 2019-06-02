import React from 'react'
import Link from 'next/link'
import { NavLink } from './Menu'

export const AnonymousNavigationBar = () => (
  <>
    <Link href="/sign-in">
      <NavLink href="">Sign in</NavLink>
    </Link>
    <Link href="/register">
      <NavLink href="">Register</NavLink>
    </Link>
  </>
)
