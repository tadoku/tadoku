import React, { useState } from 'react'
import { RankingRegistration } from '../../interfaces'
import NewLogFormModal from '../modals/NewLogFormModal'
import { NavigationBarLink } from '../../../ui/components/navigation/index'

interface Props {
  registration: RankingRegistration | undefined
  refreshRanking: () => void
}

const SubmitPages = ({ registration, refreshRanking }: Props) => {
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
          refreshRanking()
        }}
      />
    </>
  )
}

export default SubmitPages
