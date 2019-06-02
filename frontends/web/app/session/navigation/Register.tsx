import React, { useState } from 'react'
import RegisterModal from '../components/RegisterModal'
import Router from 'next/router'
import { NavigationBarLink } from '../../ui/components/navigation/index'

const Register = () => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <NavigationBarLink href="#" onClick={() => setOpen(true)}>
        Sign up
      </NavigationBarLink>
      <RegisterModal
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

export default Register
