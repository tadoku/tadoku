import React, { useState } from 'react'
import RegisterModal from '../modals/RegisterModal'
import { NavigationBarLink } from '../../../ui/components/navigation/index'

const Register = ({ refreshSession }: { refreshSession: () => void }) => {
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
          refreshSession()
        }}
      />
    </>
  )
}

export default Register
