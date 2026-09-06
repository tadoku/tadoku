import { Breadcrumb } from 'paper-ui'

export default function Example() {
  return (
    <Breadcrumb
      items={[
        { id: 'home', label: 'Home', href: '#home' },
        { id: 'contests', label: 'Contests', href: '#contests' },
        { id: 'round', label: 'August Japanese', href: '#august' },
        { id: 'entry', label: 'Reading entry 124' },
      ]}
    />
  )
}
