import React, { useState } from 'react'
import { RankingRegistration } from '../../interfaces'
import NewLogFormModal from '../modals/NewLogFormModal'
import { Button } from '../../../ui/components'

interface Props {
  registration: RankingRegistration | undefined
  refreshRanking: () => void
}

const SubmitPages = ({ registration, refreshRanking }: Props) => {
  const [open, setOpen] = useState(false)

  if (!registration) {
    return null
  }

  const disabled = registration.start > new Date()

  return (
    <>
      <Button
        onClick={() => setOpen(true)}
        icon="edit"
        plain
        disabled={disabled}
        title={
          disabled ? 'You can submit updates as soon as the contest starts' : ''
        }
      >
        Submit pages
      </Button>
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
