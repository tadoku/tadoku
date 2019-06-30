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

  const hasStarted = registration.start > new Date()
  const hasEnded = registration.end <= new Date()
  const disabled = !hasStarted || hasEnded

  let title = ''
  if (hasStarted) {
    title = 'You can submit updates as soon as the contest starts'
  }
  if (hasEnded) {
    title = 'The contest has ended, you cannot submit updates anymore'
  }

  return (
    <>
      <Button
        onClick={() => setOpen(true)}
        icon="edit"
        plain
        disabled={disabled}
        title={title}
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
