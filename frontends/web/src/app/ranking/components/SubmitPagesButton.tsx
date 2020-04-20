import React, { useState, useEffect } from 'react'
import { RankingRegistration } from '../interfaces'
import NewLogFormModal from './modals/NewLogFormModal'
import { Button } from '../../ui/components'

interface Props {
  registration: RankingRegistration | undefined
  refreshRanking: () => void
}

const SubmitPagesButton = ({ registration, refreshRanking }: Props) => {
  const [open, setOpen] = useState(false)
  const [currentDate, setCurrentDate] = useState(() => new Date())

  useEffect(() => {
    const id = setInterval(() => setCurrentDate(new Date()), 1000)
    return () => clearInterval(id)
  }, [])

  if (!registration) {
    return null
  }

  const hasStarted = registration.start <= currentDate
  const hasEnded = registration.end <= currentDate
  const disabled = !hasStarted || hasEnded

  let title = ''
  if (!hasStarted) {
    title = 'You can submit updates as soon as the contest starts'
  }
  if (hasEnded) {
    title = 'The contest has ended, you cannot submit updates anymore'
  }

  return (
    <>
      <Button
        onClick={() => setOpen(true)}
        primary
        icon="edit"
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

export default SubmitPagesButton
