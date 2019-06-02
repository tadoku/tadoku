import React, { useState } from 'react'
import RegisterModal from '../components/RegisterModal'
import { NavigationBarLink } from '../../ui/components/navigation/index'
import { refresh } from '../../router'

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
          refresh()
        }}
      />
    </>
  )
}

export default Register
