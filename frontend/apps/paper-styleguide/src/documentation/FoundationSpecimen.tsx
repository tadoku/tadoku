import { Surface } from 'paper-ui'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import wordmarkUrl from 'paper-ui/assets/brand/wordmark.svg?no-inline'
import type { CatalogDocument } from 'paper-ui/catalog'
import {
  CheckCircleIcon,
  PlusIcon,
  iconClassName,
} from 'paper-ui/icons'
import type { ReactNode } from 'react'

const COLOR_ROLES = [
  ['Canvas', '--paper-color-surface-canvas'],
  ['Paper', '--paper-color-surface-paper'],
  ['Ink', '--paper-color-text-ink'],
  ['Action', '--paper-color-action-default'],
  ['Success', '--paper-color-status-success'],
  ['Chart 1', '--paper-color-chart-1'],
] as const

const TYPE_ROLES = [
  ['Display', 'paper-type-display'],
  ['Page', 'paper-type-page'],
  ['Section', 'paper-type-section'],
  ['Component', 'paper-type-component'],
  ['Label', 'paper-type-label'],
  ['Metadata', 'paper-type-metadata'],
] as const

function PrinciplesSpecimen({ document }: { document: CatalogDocument }) {
  return (
    <div className="foundation-principles">
      <p className="foundation-specimen__lead">
        {document.guidance.whenToUse[0]}
      </p>
      <div className="foundation-principles__grid">
        <div>
          <strong>Editorial hierarchy</strong>
          <span>Reading comes before ornament.</span>
        </div>
        <div>
          <strong>Native behavior</strong>
          <span>Semantics and recognizable state come first.</span>
        </div>
        <div>
          <strong>Purposeful complexity</strong>
          <span>Add structure only when it helps a reading task.</span>
        </div>
      </div>
    </div>
  )
}

function ColorSpecimen() {
  return (
    <ul className="foundation-swatches" aria-label="Semantic color roles">
      {COLOR_ROLES.map(([label, token], index) => (
        <li key={token}>
          <span
            className={`foundation-swatch foundation-swatch--${index + 1}`}
            aria-hidden="true"
          />
          <strong>{label}</strong>
          <code>{token}</code>
        </li>
      ))}
    </ul>
  )
}

function TypographySpecimen() {
  return (
    <div className="foundation-type-scale">
      {TYPE_ROLES.map(([label, className]) => (
        <div key={className}>
          <span className={className}>Reading with Paper</span>
          <code>{className}</code>
          <span className="paper-sr-only"> — {label}</span>
        </div>
      ))}
    </div>
  )
}

function SpacingSpecimen() {
  return (
    <div className="foundation-density-grid">
      <div data-density="comfortable">
        <strong>Comfortable</strong>
        <span className="foundation-control-height paper-bg-canvas">44 px control</span>
        <code>--paper-control-height: 2.75rem</code>
      </div>
      <div data-density="compact">
        <strong>Compact</strong>
        <span className="foundation-control-height paper-bg-canvas">36 px control</span>
        <code>--paper-control-height: 2.25rem</code>
      </div>
    </div>
  )
}

function LayoutSpecimen({ document }: { document: CatalogDocument }) {
  return (
    <div className="foundation-layout">
      <p className="foundation-specimen__lead">{document.guidance.content[0]}</p>
      <div className="foundation-layout__flow">
        <div>
          <strong>Readable measure</strong>
          <p>Prose stays narrow enough to scan while the page remains fluid.</p>
        </div>
        <div className="foundation-layout__cluster" aria-label="Wrapping controls">
          <span className="paper-bg-canvas">Month</span>
          <span className="paper-bg-canvas">Language</span>
          <span className="paper-bg-canvas">Contest</span>
        </div>
      </div>
      <div
        className="foundation-layout__overflow"
        role="region"
        aria-label="Horizontal data example"
        tabIndex={0}
      >
        Labelled overflow keeps wide data reachable without widening the page.
      </div>
    </div>
  )
}

function ShapeSpecimen() {
  return (
    <div className="foundation-shapes">
      <div className="paper-accent-rail paper-bg-canvas">
        <strong>Accent rail</strong>
        <code>paper-accent-rail</code>
      </div>
      <dl>
        <div><dt>Static rule</dt><dd><code>--paper-border-static-width</code></dd></div>
        <div><dt>Field edge</dt><dd><code>--paper-border-field-edge-width</code></dd></div>
        <div><dt>Action edge</dt><dd><code>--paper-border-action-edge-width</code></dd></div>
      </dl>
    </div>
  )
}

function ElevationSpecimen() {
  return (
    <div className="foundation-elevations">
      {(['flat', 'floating', 'showcase'] as const).map((elevation) => (
        <Surface key={elevation} elevation={elevation}>
          <strong>{elevation}</strong>
          <code>{`paper-elevation-${elevation}`}</code>
        </Surface>
      ))}
    </div>
  )
}

function IconographySpecimen() {
  return (
    <div className="foundation-icons">
      <div>
        <PlusIcon className={iconClassName('compact')} aria-hidden="true" />
        <strong>PlusIcon</strong>
        <code>paper-icon-compact</code>
        <span>Outline · action</span>
      </div>
      <div>
        <CheckCircleIcon className={iconClassName()} aria-hidden="true" />
        <strong>CheckCircleIcon</strong>
        <span>Solid · status</span>
      </div>
    </div>
  )
}

function MotionSpecimen() {
  const roles = [
    ['Quick feedback', '--paper-motion-quick', '120ms', 'quick'],
    ['Standard transition', '--paper-motion-standard', '180ms', 'standard'],
    ['Deliberate movement', '--paper-motion-deliberate', '240ms', 'deliberate'],
  ] as const

  return (
    <div className="foundation-motion">
      {roles.map(([label, token, duration, speed]) => (
        <div key={token}>
          <strong>{label}</strong>
          <span className={`foundation-motion__track foundation-motion__track--${speed} paper-bg-canvas`} aria-hidden="true">
            <span />
          </span>
          <code>{token}</code>
          <span>{duration}</span>
        </div>
      ))}
    </div>
  )
}

function BrandSpecimen() {
  return (
    <div className="foundation-brand">
      <figure className="paper-bg-paper" data-theme="light">
        <img src={cutMeterUrl} alt="Cut Meter" />
        <figcaption>Canonical monochrome mark</figcaption>
      </figure>
      <figure className="paper-bg-paper" data-theme="light">
        <img src={wordmarkUrl} alt="Tadoku" />
        <figcaption>Canonical wordmark</figcaption>
      </figure>
    </div>
  )
}

const SPECIMENS: Readonly<Record<string, (document: CatalogDocument) => ReactNode>> = {
  'foundation.principles': (document) => <PrinciplesSpecimen document={document} />,
  'foundation.color': () => <ColorSpecimen />,
  'foundation.typography': () => <TypographySpecimen />,
  'foundation.spacing-and-density': () => <SpacingSpecimen />,
  'foundation.layout': (document) => <LayoutSpecimen document={document} />,
  'foundation.shape-and-borders': () => <ShapeSpecimen />,
  'foundation.elevation': () => <ElevationSpecimen />,
  'foundation.iconography': () => <IconographySpecimen />,
  'foundation.motion': () => <MotionSpecimen />,
  'foundation.brand': () => <BrandSpecimen />,
}

export function FoundationSpecimen({ document }: { document: CatalogDocument }) {
  const renderSpecimen = SPECIMENS[document.id]

  return (
    <Surface
      as="section"
      aria-label={`${document.name} specimen`}
      className="foundation-specimen"
      data-foundation-specimen={document.id}
    >
      {renderSpecimen?.(document)}
    </Surface>
  )
}
