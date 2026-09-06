import { ButtonGroup } from 'paper-ui'
import { useState } from 'react'

export default function EntryActionsFixture() {
  const [message, setMessage] = useState('Choose an action to see its result.')
  return (
    <div className="paper-stack">
      <ButtonGroup
        label="Entry actions"
        actions={[
          {
            id: 'view',
            label: 'View log',
            href: '#reading-summary',
            variant: 'outline',
          },
          {
            id: 'edit',
            label: 'Edit log',
            onSelect: () => setMessage('Editing the local preview entry.'),
          },
          {
            id: 'delete',
            label: 'Delete log',
            variant: 'destructive',
            disabled: true,
          },
        ]}
      />
      <p role="status">{message}</p>
      <p id="reading-summary">
        Reading log: 48 pages of Japanese fiction. Deletion is unavailable for
        this submitted entry.
      </p>
    </div>
  )
}
