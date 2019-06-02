import React, { useState } from 'react'
import LogInModal from '../components/LogInModal'
import Router from 'next/router'
import { NavigationBarLink } from '../../ui/components/navigation/index'

const LogIn = () => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <NavigationBarLink href="#" onClick={() => setOpen(true)}>
        Log in
      </NavigationBarLink>
      <LogInModal
        isOpen={open}
        onCancel={() => setOpen(false)}
        onSuccess={() => {
          setOpen(false)
          if (Router.asPath) {
            Router.push('/ranking')
          }
        }}
      />
    </>
  )
}

export default LogIn
