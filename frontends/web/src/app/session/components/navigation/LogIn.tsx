import React, { useState } from 'react'
import LogInModal from '../modals/LogInModal'
import { Button } from '@app/ui/components'

const LogIn = ({ refreshSession }: { refreshSession: () => void }) => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button onClick={() => setOpen(true)} plain>
        Log in
      </Button>
      <LogInModal
        isOpen={open}
        onCancel={() => setOpen(false)}
        onSuccess={() => {
          refreshSession()
        }}
      />
    </>
  )
}

export default LogIn
