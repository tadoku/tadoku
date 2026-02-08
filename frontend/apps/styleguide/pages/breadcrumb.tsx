import { Showcase } from '@components/Showcase'

import BreadcrumbExample from '@examples/breadcrumb/example'
import exampleCode from '@examples/breadcrumb/example.tsx?raw'

export default function BreadcrumbPage() {
  return (
    <>
      <h1 className="title mb-8">Breadcrumb</h1>

      <Showcase title="Example" code={exampleCode}>
        <BreadcrumbExample />
      </Showcase>
    </>
  )
}
