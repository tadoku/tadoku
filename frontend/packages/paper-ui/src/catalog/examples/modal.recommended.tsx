import { Modal } from 'paper-ui'
import { useState } from 'react'

export default function DeleteLogFixture() {
  const [deleted, setDeleted] = useState(false)
  return (
    <>
      <Modal
        triggerLabel="Review deletion"
        triggerVariant="destructive"
        title="Delete this reading log?"
        description="This demo only changes its local example state."
        closeLabel="Keep log"
        action={{
          label: 'Delete log',
          variant: 'destructive',
          onAction: () => setDeleted(true),
        }}
      >
        <p>August Japanese reading · 1,240 pages</p>
      </Modal>
      <p role="status">
        {deleted ? 'Example log deleted' : 'Example log is available'}
      </p>
    </>
  )
}
