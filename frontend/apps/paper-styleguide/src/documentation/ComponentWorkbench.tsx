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
      className="component-workbench"
      aria-label={`${document.name} examples`}
    >
      <Tabs.Root defaultValue="preview">
        <div className="component-workbench__tabs">
          <Tabs.List aria-label="Example views">
            {VIEWS.map(candidate => (
              <Tabs.Tab key={candidate} value={candidate}>
                {label(candidate)}
              </Tabs.Tab>
            ))}
          </Tabs.List>
        </div>

        <Tabs.Panel value="preview" className="component-workbench__panel">
          <ExampleCanvas
            fixture={fixture}
            fixtures={fixtures}
            onFixtureChange={changeFixture}
          />
        </Tabs.Panel>

        <Tabs.Panel value="code" className="component-workbench__panel">
          <div className="code-view">
            <div className="code-view__heading">
              <div>
                <h3>{fixture?.name ?? 'Example'}</h3>
                <p>{fixture?.description}</p>
              </div>
              <div className="code-copy">
                <Button
                  variant="outline"
                  className="code-copy__button"
                  disabled={!fixture?.code}
                  onClick={copyCode}
                >
                  {copyState === 'copied' ? 'Copied' : 'Copy code'}
                </Button>
                <span
                  className="code-copy__status"
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

        <Tabs.Panel value="api" className="component-workbench__panel">
          <div className="workbench-api">
            {document.api.props?.length ? (
              <Table
                caption={`${document.name} props`}
                tableClassName="workbench-props-table"
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
                        <span className="workbench-prop-type">
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
                        <p className="workbench-prop-default">
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
            <div className="workbench-reference-grid">
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
          className="component-workbench__panel"
        >
          <div className="workbench-reference-grid">
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
