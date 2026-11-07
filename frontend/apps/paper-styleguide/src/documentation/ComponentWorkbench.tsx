import * as stylex from '@stylexjs/stylex'
import { useState } from 'react'
import { Button, Surface, Table, Tabs } from 'paper-ui'
import { size } from 'paper-ui/tokens.stylex'
import type { CatalogDocument, CatalogFixture } from 'paper-ui/catalog'
import { ExampleCanvas } from './ExampleCanvas'
import { CodeBlock } from './CodeBlock'
import { shellStyles } from '../styles/shell.styles'
import { descendantStyles } from '../descendants.styles'

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
      xstyle={[shellStyles.componentWorkbench]}
      aria-label={`${document.name} examples`}
    >
      <Tabs.Root defaultValue="preview">
        <div {...stylex.props(shellStyles.componentWorkbench__tabs)}>
          <Tabs.List aria-label="Example views">
            {VIEWS.map(candidate => (
              <Tabs.Tab key={candidate} value={candidate}>
                {label(candidate)}
              </Tabs.Tab>
            ))}
          </Tabs.List>
        </div>

        <Tabs.Panel value="preview" {...stylex.props(shellStyles.componentWorkbench__panel)}>
          <ExampleCanvas
            fixture={fixture}
            fixtures={fixtures}
            onFixtureChange={changeFixture}
          />
        </Tabs.Panel>

        <Tabs.Panel value="code" {...stylex.props(shellStyles.componentWorkbench__panel)}>
          <div {...stylex.props(shellStyles.codeView)}>
            <div {...stylex.props(shellStyles.codeView__heading)}>
              <div>
                <h3 {...stylex.props(descendantStyles.codeViewU_uH3)}>{fixture?.name ?? 'Example'}</h3>
                <p>{fixture?.description}</p>
              </div>
              <div {...stylex.props(shellStyles.codeCopy)}>
                <Button
                  variant="outline"
                  disabled={!fixture?.code}
                  onClick={copyCode}
                >
                  {copyState === 'copied' ? 'Copied' : 'Copy code'}
                </Button>
                <span
                  {...stylex.props(shellStyles.codeCopy__status)}
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

        <Tabs.Panel value="api" {...stylex.props(shellStyles.componentWorkbench__panel)}>
          <div {...stylex.props(shellStyles.workbenchApi)}>
            {document.api.props?.length ? (
              <Table
                caption={`${document.name} props`}
                tableXstyle={shellStyles.workbenchPropsTable}
                rows={document.api.props}
                getRowKey={prop => prop.name}
                minWidth={size[544]}
                columns={[
                  {
                    id: 'name',
                    header: 'Prop / Type',
                    rowHeader: true,
                    width: '40%',
                    cell: prop => (
                      <>
                        <code {...stylex.props(descendantStyles.workbenchApiU_uCode)}>{prop.name}</code>
                        <span {...stylex.props(shellStyles.workbenchPropType)}>
                          <code {...stylex.props(descendantStyles.workbenchApiU_uCode)}>{prop.type}</code>
                        </span>
                      </>
                    ),
                  },
                  {
                    id: 'description',
                    header: 'Usage',
                    cell: prop => (
                      <>
                        <p {...stylex.props(shellStyles.workbenchPropDefault)}>
                          {prop.required ? (
                            <strong>Required</strong>
                          ) : (
                            <>
                              Default: <code {...stylex.props(descendantStyles.workbenchApiU_uCode)}>{prop.defaultValue ?? '—'}</code>
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
            {document.api.styleExports.length ? (
              <Table
                caption="CSS classes and helpers"
                rows={document.api.styleExports}
                getRowKey={name => name}
                minWidth={size[400]}
                columns={[
                  {
                    id: 'name',
                    header: 'Name',
                    rowHeader: true,
                    cell: name => <code {...stylex.props(descendantStyles.workbenchApiU_uCode)}>{name}</code>,
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
            <div {...stylex.props(shellStyles.workbenchReferenceGrid, descendantStyles.workbenchApiU_uWorkbenchReferenceGrid)}>
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
          {...stylex.props(shellStyles.componentWorkbench__panel)}
        >
          <div {...stylex.props(shellStyles.workbenchReferenceGrid)}>
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
