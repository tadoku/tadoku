import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import ExampleTable from '@examples/tables/example'
import exampleCode from '@examples/tables/example.tsx?raw'

import TableWithActionMenu from '@examples/tables/with-action-menu'
import withActionMenuCode from '@examples/tables/with-action-menu.tsx?raw'

export default function Page() {
  return (
    <>
      <h1 className="title mb-8">Table</h1>

      <Showcase title="Example" code={exampleCode}>
        <ExampleTable />
      </Showcase>

      <Separator />

      <Showcase title="Clickable Rows with ActionMenu" code={withActionMenuCode}>
        <TableWithActionMenu />
      </Showcase>
    </>
  )
}
