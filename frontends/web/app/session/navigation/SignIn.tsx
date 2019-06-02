import React, { useState } from 'react'
import SignInModal from '../components/SignInModal'
import Router from 'next/router'
import { NavigationBarLink } from '../../ui/components/navigation/index'

export const SignIn = () => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <NavigationBarLink href="#" onClick={() => setOpen(true)}>
        Sign in
      </NavigationBarLink>
      <SignInModal
        isOpen={open}
        onCancel={() => setOpen(false)}
        onSuccess={() => {
          setOpen(false)
          if (Router.asPath) {
            Router.push(Router.asPath)
          }
        }}
      />
    </>
  )
}
