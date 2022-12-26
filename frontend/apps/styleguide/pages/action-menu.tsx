import { ActionMenu } from 'ui/components/ActionMenu'
import { CodeBlock, Preview, Title } from '@components/example'
import {
  EllipsisVerticalIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/react/20/solid'
import { toast } from 'react-toastify'

export default function Page() {
  return (
    <>
      <h1 className="title mb-8">Action Menu</h1>

      <Title>Example</Title>
      <Preview>
        <ExampleMenu />
      </Preview>
      <CodeBlock
        code={`import { ActionMenu } from '@components/ActionMenu'
import {
  EllipsisVerticalIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/react/20/solid'

function ExampleMenu() {
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
}`}
      />
    </>
  )
}

function ExampleMenu() {
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
