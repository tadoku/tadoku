import { useState } from 'react'
import { Button, Surface, Tabs } from 'paper-ui'
import type { CatalogDocument, CatalogFixture } from 'paper-ui/catalog'
import { ExampleCanvas } from './ExampleCanvas'

const VIEWS = ['preview', 'code', 'api', 'accessibility'] as const
type View = (typeof VIEWS)[number]
type CopyState = 'idle' | 'copied' | 'error'

function label(view: View): string {
  if (view === 'api') return 'API / Props'
  return `${view.charAt(0).toUpperCase()}${view.slice(1)}`
}

function ApiList({ title, items }: { title: string; items: readonly string[] }) {
  return (
    <section>
      <h4>{title}</h4>
      {items.length ? (
        <ul>
          {items.map((item) => (
            <li key={item}>
              <code>{item}</code>
            </li>
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
    fixtures.find((candidate) => candidate.id === fixtureId) ?? fixtures[0]

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
            {VIEWS.map((candidate) => (
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
            <pre tabIndex={0}>
              <code>
                {fixture?.code ?? 'No copyable example is registered.'}
              </code>
            </pre>
          </div>
        </Tabs.Panel>

        <Tabs.Panel value="api" className="component-workbench__panel">
          <div className="workbench-reference-grid">
            <ApiList title="React" items={document.api.react} />
            <ApiList title="CSS and recipes" items={document.api.cssClasses} />
            <ApiList title="Public types" items={document.api.publicTypes} />
            <ApiList title="Defaults" items={document.api.defaults} />
            <ApiList
              title="Invalid combinations"
              items={document.api.invalidCombinations}
            />
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
