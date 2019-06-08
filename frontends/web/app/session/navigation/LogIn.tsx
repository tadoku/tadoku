import React, { useState } from 'react'
import LogInModal from '../components/modals/LogInModal'
import { NavigationBarLink } from '../../ui/components/navigation/index'

const LogIn = ({ refreshSession }: { refreshSession: () => void }) => {
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
          refreshSession()
        }}
      />
    </>
  )
}

export default LogIn
