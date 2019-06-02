import React, { useState } from 'react'
import LogInModal from '../components/LogInModal'
import { NavigationBarLink } from '../../ui/components/navigation/index'
import { refresh } from '../../router'

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
          refresh()
        }}
      />
    </>
  )
}

export default LogIn
