import Breadcrumb from '@components/Breadcrumb'
import { CodeBlock, Preview, Title } from '@components/example'
import { HomeIcon } from '@heroicons/react/20/solid'

export default function BreadcrumbExample() {
  return (
    <>
      <h1 className="title mb-8">Breadcrumb</h1>

      <Title>Example</Title>
      <Preview>
        <Breadcrumb
          links={[
            { label: 'Home', href: '/', IconComponent: HomeIcon },
            { label: 'Contests', href: '/contests' },
            { label: '2022 Round 6', href: '/contests/20' },
            { label: 'antonve', href: '/contests/20/1' },
          ]}
        />
      </Preview>
      <CodeBlock
        code={`import Breadcrumb from '@components/Breadcrumb'
import { HomeIcon } from '@heroicons/react/20/solid'

const Example = () => (
  <Breadcrumb
    links={[
      { label: 'Home', href: '/', IconComponent: HomeIcon },
      { label: 'Contests', href: '/contests' },
      { label: '2022 Round 6', href: '/contests/20' },
      { label: 'antonve', href: '/contests/20/1' },
    ]}
  />
)`}
      />
    </>
  )
}
