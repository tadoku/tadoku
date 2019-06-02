import React, { useState } from 'react'
import { RankingRegistration } from '../interfaces'
import NewLogFormModal from '../components/NewLogFormModal'
import { NavigationBarLink } from '../../ui/components/navigation/index'
import { refresh } from '../../router'

const SubmitPages = ({
  registration,
}: {
  registration: RankingRegistration | undefined
}) => {
  const [open, setOpen] = useState(false)

  if (!registration) {
    return null
  }

  return (
    <>
      <NavigationBarLink href="#" onClick={() => setOpen(true)}>
        Submit pages
      </NavigationBarLink>
      <NewLogFormModal
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

export default SubmitPages
