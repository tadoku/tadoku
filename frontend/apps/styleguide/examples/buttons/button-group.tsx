import { ButtonGroup } from 'ui'

export default function ButtonGroupExample() {
  return (
    <ButtonGroup
      actions={[
        {
          href: '#',
          onClick: () => console.log('pressed'),
          label: 'Primary',
          style: 'primary',
        },
        { href: '#', label: 'Secondary', style: 'secondary' },
        { href: '#', label: 'Tertiary (default)' },
        { href: '#', label: 'Danger', style: 'danger' },
        { href: '#', label: 'Ghost', style: 'ghost' },
      ]}
    />
  )
}
