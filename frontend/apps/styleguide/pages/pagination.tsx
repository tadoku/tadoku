import { Showcase } from '@components/Showcase'

import PaginationExample from '@examples/pagination/example'
import exampleCode from '@examples/pagination/example.tsx?raw'

export default function Page() {
  return (
    <>
      <h1 className="title mb-8">Pagination</h1>

      <Showcase title="Example" code={exampleCode}>
        <PaginationExample />
      </Showcase>
    </>
  )
}
