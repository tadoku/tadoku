import React, { useState } from 'react'
import { RankingRegistration } from '../interfaces'
import NewLogFormModal from '../components/NewLogFormModal'
import Router from 'next/router'
import { NavigationBarLink } from '../../ui/components/navigation/index'

export const SubmitPages = ({
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
          if (Router.asPath) {
            Router.push(Router.asPath)
          }
        }}
      />
    </>
  )
}
