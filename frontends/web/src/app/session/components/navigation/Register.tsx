import React, { useState } from 'react'
import RegisterModal from '../modals/RegisterModal'
import { Button } from '@app/ui/components'

const Register = ({ refreshSession }: { refreshSession: () => void }) => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button onClick={() => setOpen(true)} plain>
        Sign up
      </Button>
      <RegisterModal
        isOpen={open}
        onCancel={() => setOpen(false)}
        onSuccess={() => {
          refreshSession()
        }}
      />
    </>
  )
}

export default Register
