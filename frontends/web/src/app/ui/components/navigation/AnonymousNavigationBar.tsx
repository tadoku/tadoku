import React from 'react'
import LogInLink from '../../../session/components/navigation/LogIn'
import RegisterLink from '../../../session/components/navigation/Register'
import Link from 'next/link'
import { ButtonLink } from '..'
import LinkContainer from './LinkContainer'

export const AnonymousNavigationBar = ({
  refreshSession,
}: {
  refreshSession: () => void
}) => (
  <LinkContainer>
    <Link href="/blog" passHref>
      <ButtonLink plain>Blog</ButtonLink>
    </Link>
    <Link href="/ranking" passHref>
      <ButtonLink plain>Ranking</ButtonLink>
    </Link>
    <Link href="/manual" passHref>
      <ButtonLink plain>Manual</ButtonLink>
    </Link>
    <LogInLink refreshSession={refreshSession} />
    <RegisterLink refreshSession={refreshSession} />
  </LinkContainer>
)
