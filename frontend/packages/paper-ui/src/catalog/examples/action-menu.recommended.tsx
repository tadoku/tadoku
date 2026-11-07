import * as stylex from '@stylexjs/stylex'
import { ActionMenu, layout } from 'paper-ui'
import { UserCircleIcon } from 'paper-ui/icons'
import { useState } from 'react'

export default function ActionMenuFixture() {
  const [result, setResult] = useState('Choose a log action.')
  const items = [
    {
      id: 'edit',
      label: 'Edit log',
      onSelect: () => setResult('Edit log selected'),
    },
    {
      id: 'duplicate',
      label: 'Duplicate log',
      disabled: true,
      onSelect: () => setResult('Duplicate log selected'),
    },
    {
      id: 'delete',
      label: 'Delete log',
      destructive: true,
      onSelect: () =>
        setResult(
          'Delete log selected — request confirmation before deleting data.',
        ),
    },
  ]
  return (
    <div {...stylex.props(layout.stack)}>
      <div {...stylex.props(layout.cluster)}>
        <ActionMenu label="Log actions" items={items} />
      </div>
      <p>
        Use a labelled trigger when the menu needs to explain its scope. In a
        compact row, an ellipsis can use the row title as context.
      </p>
      <div {...stylex.props(layout.cluster)}>
        <span>August Japanese reading log</span>
        <ActionMenu
          label="Actions for August Japanese reading log"
          items={items}
          iconOnly
          triggerVariant="ghost"
        />
      </div>
      <div {...stylex.props(layout.cluster)}>
        <span>Account menu</span>
        <ActionMenu label="Anton account menu" iconOnly triggerIcon={<UserCircleIcon />} triggerVariant="ghost"
          items={[{ id: 'profile', label: 'My profile', onSelect: () => setResult('My profile selected') }]} />
      </div>
      <p role="status">{result}</p>
      <p>Duplication is unavailable for this archived example.</p>
    </div>
  )
}
