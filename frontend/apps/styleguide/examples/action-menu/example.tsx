import { ActionMenu } from 'ui'
import {
  EllipsisVerticalIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/react/20/solid'
import { toast } from 'react-toastify'

export default function ActionMenuExample() {
  return (
    <ActionMenu
      links={[
        { label: 'Edit', href: '#', IconComponent: PencilIcon },
        {
          label: 'Delete',
          href: '#',
          IconComponent: TrashIcon,
          type: 'danger',
          onClick: () => toast.warn('Deleted...'),
        },
      ]}
    >
      <EllipsisVerticalIcon className="w-4 h-5" />
    </ActionMenu>
  )
}
