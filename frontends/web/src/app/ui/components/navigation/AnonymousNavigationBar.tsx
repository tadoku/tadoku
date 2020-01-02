import React from 'react'
import LogInLink from '../../../session/components/navigation/LogIn'
import RegisterLink from '../../../session/components/navigation/Register'
import Link from 'next/link'
import { Button } from '..'

export const AnonymousNavigationBar = ({
  refreshSession,
}: {
  refreshSession: () => void
}) => (
  <>
    <Link href="/blog">
      <a href="">
        <Button plain>Blog</Button>
      </a>
    </Link>
    <Link href="/ranking">
      <a href="">
        <Button plain>Ranking</Button>
      </a>
    </Link>
    <Link href="/manual">
      <a href="">
        <Button plain>Manual</Button>
      </a>
    </Link>
    <Link href="/about">
      <a href="">
        <Button plain>About</Button>
      </a>
    </Link>
    <LogInLink refreshSession={refreshSession} />
    <RegisterLink refreshSession={refreshSession} />
  </>
)
