import React from 'react'
import Modal from '../../../ui/components/Modal'
import JoinContestForm from '../JoinContestForm'
import { Contest } from '../../../contest/interfaces'

const JoinContestModal = ({
  contest,
  isOpen,
  onSuccess,
  onCancel,
}: {
  contest: Contest
  isOpen: boolean
  onSuccess: () => void
  onCancel: () => void
}) => {
  return (
    <Modal
      isOpen={isOpen}
      onRequestClose={onCancel}
      contentLabel="Join contest"
    >
      <JoinContestForm
        contest={contest}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    </Modal>
  )
}

export default JoinContestModal
