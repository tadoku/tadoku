import { useRef, useState, type KeyboardEvent } from 'react'
import type { CatalogDocument, CatalogFixture } from 'paper-ui/catalog'
import { ExampleCanvas } from './ExampleCanvas'

const VIEWS = ['preview', 'code', 'api', 'accessibility'] as const
type View = (typeof VIEWS)[number]

function label(view: View): string {
  if (view === 'api') return 'API / Props'
  return `${view.charAt(0).toUpperCase()}${view.slice(1)}`
}

function ApiList({ title, items }: { title: string; items: readonly string[] }) {
  return (
    <section>
      <h4>{title}</h4>
      {items.length ? (
        <ul>{items.map((item) => <li key={item}><code>{item}</code></li>)}</ul>
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
  const [view, setView] = useState<View>('preview')
  const [fixtureId, setFixtureId] = useState(fixtures[0]?.id ?? '')
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([])
  const fixture = fixtures.find((candidate) => candidate.id === fixtureId) ?? fixtures[0]

  function moveTab(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    event.preventDefault()
    const nextIndex = event.key === 'Home'
      ? 0
      : event.key === 'End'
        ? VIEWS.length - 1
        : (index + (event.key === 'ArrowRight' ? 1 : -1) + VIEWS.length) % VIEWS.length
    const nextView = VIEWS[nextIndex]
    setView(nextView)
    tabRefs.current[nextIndex]?.focus()
  }

  return (
    <section className="component-workbench" aria-label={`${document.name} examples`}>
      <div className="component-workbench__toolbar">
        <div className="component-workbench__tabs" role="tablist" aria-label="Example views">
          {VIEWS.map((candidate, index) => (
            <button
              key={candidate}
              ref={(node) => { tabRefs.current[index] = node }}
              id={`workbench-tab-${candidate}`}
              type="button"
              role="tab"
              aria-selected={view === candidate}
              aria-controls={`workbench-panel-${candidate}`}
              tabIndex={view === candidate ? 0 : -1}
              onKeyDown={(event) => moveTab(event, index)}
              onClick={() => setView(candidate)}
            >
              {label(candidate)}
            </button>
          ))}
        </div>
        {fixtures.length > 1 ? (
          <label className="component-workbench__fixture-select">
            <span>Fixture</span>
            <select value={fixture?.id} onChange={(event) => setFixtureId(event.target.value)}>
              {fixtures.map((candidate) => (
                <option key={candidate.id} value={candidate.id}>{candidate.name}</option>
              ))}
            </select>
          </label>
        ) : null}
      </div>

      <div
        id={`workbench-panel-${view}`}
        role="tabpanel"
        aria-labelledby={`workbench-tab-${view}`}
        className="component-workbench__panel"
      >
        {view === 'preview' ? <ExampleCanvas fixture={fixture} /> : null}
        {view === 'code' ? (
          <div className="code-view">
            <div>
              <h3>{fixture?.name ?? 'Example'}</h3>
              <p>{fixture?.description}</p>
            </div>
            <pre tabIndex={0}><code>{fixture?.code ?? 'No copyable example is registered.'}</code></pre>
          </div>
        ) : null}
        {view === 'api' ? (
          <div className="workbench-reference-grid">
            <ApiList title="React" items={document.api.react} />
            <ApiList title="CSS and recipes" items={document.api.cssClasses} />
            <ApiList title="Public types" items={document.api.publicTypes} />
            <ApiList title="Defaults" items={document.api.defaults} />
            <ApiList title="Invalid combinations" items={document.api.invalidCombinations} />
          </div>
        ) : null}
        {view === 'accessibility' ? (
          <div className="workbench-reference-grid">
            <ApiList title="Requirements" items={document.accessibility.requirements} />
            <ApiList title="Keyboard" items={document.accessibility.keyboard} />
            <ApiList title="Known constraints" items={document.accessibility.knownConstraints} />
          </div>
        ) : null}
      </div>
    </section>
  )
}
