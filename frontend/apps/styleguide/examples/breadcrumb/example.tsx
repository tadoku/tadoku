import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'

export default function BreadcrumbExample() {
  return (
    <Breadcrumb
      links={[
        { label: 'Home', href: '#home', IconComponent: HomeIcon },
        { label: 'Contests', href: '#contests' },
        { label: '2022 Round 6', href: '#contest-20' },
        { label: 'antonve', href: '#user' },
      ]}
    />
  )
}
