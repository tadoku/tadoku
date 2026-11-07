import { useState } from 'react'
import { Button, Surface, Table, Tabs } from 'paper-ui'
import type { CatalogDocument, CatalogFixture } from 'paper-ui/catalog'
import { ExampleCanvas } from './ExampleCanvas'
import { CodeBlock } from './CodeBlock'

const VIEWS = ['preview', 'code', 'api', 'accessibility'] as const
type View = (typeof VIEWS)[number]
type CopyState = 'idle' | 'copied' | 'error'

function label(view: View): string {
  if (view === 'api') return 'API / Props'
  return `${view.charAt(0).toUpperCase()}${view.slice(1)}`
}

function ApiList({
  title,
  items,
  code = false,
}: {
  title: string
  items: readonly string[]
  code?: boolean
}) {
  if (!items.length) return null
  return (
    <section>
      <h4>{title}</h4>
      {items.length ? (
        <ul>
          {items.map(item => (
            <li key={item}>{code ? <code>{item}</code> : item}</li>
          ))}
        </ul>
      ) : (
        <p>None.</p>
      )}
    </section>
  )
}

export function ComponentWorkbench({
  document,
  fixtures,
}: {
  document: CatalogDocument
  fixtures: readonly CatalogFixture[]
}) {
  const [fixtureId, setFixtureId] = useState(fixtures[0]?.id ?? '')
  const [copyState, setCopyState] = useState<CopyState>('idle')
  const fixture =
    fixtures.find(candidate => candidate.id === fixtureId) ?? fixtures[0]

  function changeFixture(nextFixture: CatalogFixture) {
    if (nextFixture.id === fixtureId) return
    setFixtureId(nextFixture.id)
    setCopyState('idle')
  }

  async function copyCode() {
    const code = fixture?.code
    if (!code || !navigator.clipboard?.writeText) {
      setCopyState('error')
      return
    }

    try {
      await navigator.clipboard.writeText(code)
      setCopyState('copied')
    } catch {
      setCopyState('error')
    }
  }

  return (
    <Surface
      as="section"
      className="component-workbench overflow-clip p-0"
      aria-label={`${document.name} examples`}
    >
      <Tabs.Root defaultValue="preview">
        <div className="component-workbench__tabs min-w-0 px-3">
          <Tabs.List aria-label="Example views">
            {VIEWS.map(candidate => (
              <Tabs.Tab key={candidate} value={candidate}>
                {label(candidate)}
              </Tabs.Tab>
            ))}
          </Tabs.List>
        </div>

        <Tabs.Panel value="preview" className="component-workbench__panel min-w-0 p-0">
          <ExampleCanvas
            fixture={fixture}
            fixtures={fixtures}
            onFixtureChange={changeFixture}
          />
        </Tabs.Panel>

        <Tabs.Panel value="code" className="component-workbench__panel min-w-0 p-0">
          <div className="code-view p-5 [&_h3]:mt-0 [&_pre]:mb-0 [&_pre]:max-h-code [&_pre]:overflow-auto [&_pre]:border-s-rail [&_pre]:border-action [&_pre]:bg-canvas [&_pre]:p-4 [&_pre]:font-mono [&_pre]:text-ink">
            <div className="flex items-start justify-between gap-4 max-tablet:flex-col max-tablet:items-stretch">
              <div>
                <h3>{fixture?.name ?? 'Example'}</h3>
                <p>{fixture?.description}</p>
              </div>
              <div className="grid min-w-48 justify-items-end gap-1.4 max-tablet:min-w-0 max-tablet:justify-items-start">
                <Button
                  variant="outline"
                  className="code-copy__button"
                  disabled={!fixture?.code}
                  onClick={copyCode}
                >
                  {copyState === 'copied' ? 'Copied' : 'Copy code'}
                </Button>
                <span
                  className="min-h-metadata-line text-end text-metadata text-muted"
                  role="status"
                  aria-live="polite"
                >
                  {copyState === 'copied'
                    ? 'Code copied to clipboard.'
                    : copyState === 'error'
                    ? 'Copy failed. Select the code and copy it manually.'
                    : ''}
                </span>
              </div>
            </div>
            <CodeBlock
              code={fixture?.code ?? 'No copyable example is registered.'}
            />
          </div>
        </Tabs.Panel>

        <Tabs.Panel value="api" className="component-workbench__panel min-w-0 p-0">
          <div className="grid gap-8 p-5 [&_code]:wrap-normal [&_code]:break-normal [&_.workbench-reference-grid]:p-0">
            {document.api.props?.length ? (
              <Table
                caption={`${document.name} props`}
                tableClassName="table-fixed"
                rows={document.api.props}
                getRowKey={prop => prop.name}
                minWidth="34rem"
                columns={[
                  {
                    id: 'name',
                    header: 'Prop / Type',
                    rowHeader: true,
                    width: '40%',
                    cell: prop => (
                      <>
                        <code>{prop.name}</code>
                        <span className="mt-1.6 block text-fine font-regular">
                          <code>{prop.type}</code>
                        </span>
                      </>
                    ),
                  },
                  {
                    id: 'description',
                    header: 'Usage',
                    cell: prop => (
                      <>
                        <p className="mb-2 mt-0 text-fine">
                          {prop.required ? (
                            <strong>Required</strong>
                          ) : (
                            <>
                              Default: <code>{prop.defaultValue ?? '—'}</code>
                            </>
                          )}
                        </p>
                        {prop.description}
                      </>
                    ),
                  },
                ]}
              />
            ) : null}
            {document.api.cssClasses.length ? (
              <Table
                caption="CSS classes and helpers"
                rows={document.api.cssClasses}
                getRowKey={name => name}
                minWidth="25rem"
                columns={[
                  {
                    id: 'name',
                    header: 'Name',
                    rowHeader: true,
                    cell: name => <code>{name}</code>,
                  },
                  {
                    id: 'kind',
                    header: 'Use',
                    cell: name =>
                      name.includes('(')
                        ? 'Optional JavaScript helper that returns CSS class names.'
                        : 'CSS class: use in className or class.',
                  },
                ]}
              />
            ) : null}
            <div className="grid grid-cols-1 gap-6 p-5 [&_h4]:mb-2 [&_h4]:mt-0 [&_section]:border-0 [&_section]:p-0 [&_ul]:m-0 [&_ul]:ps-5 [&_li+li]:mt-1.6 [&_li]:wrap-anywhere [&_code]:wrap-normal">
              <ApiList title="React exports" items={document.api.react} code />
              <ApiList
                title="Public types"
                items={document.api.publicTypes}
                code
              />
              <ApiList title="Defaults" items={document.api.defaults} />
              <ApiList
                title="Invalid combinations"
                items={document.api.invalidCombinations}
              />
            </div>
          </div>
        </Tabs.Panel>

        <Tabs.Panel
          value="accessibility"
          className="component-workbench__panel min-w-0 p-0"
        >
          <div className="grid grid-cols-1 gap-6 p-5 [&_h4]:mb-2 [&_h4]:mt-0 [&_section]:border-0 [&_section]:p-0 [&_ul]:m-0 [&_ul]:ps-5 [&_li+li]:mt-1.6 [&_li]:wrap-anywhere [&_code]:wrap-normal">
            <ApiList
              title="Requirements"
              items={document.accessibility.requirements}
            />
            <ApiList title="Keyboard" items={document.accessibility.keyboard} />
            <ApiList
              title="Known constraints"
              items={document.accessibility.knownConstraints}
            />
          </div>
        </Tabs.Panel>
      </Tabs.Root>
    </Surface>
  )
}
