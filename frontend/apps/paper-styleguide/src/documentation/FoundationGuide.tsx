import { useState } from 'react'
import { Button } from 'paper-ui'
import type { CatalogDocument } from 'paper-ui/catalog'
import { CodeBlock } from './CodeBlock'
import './foundation-guide.css'

export function GuideCode({ code, label = 'Complete example' }: { code: string; label?: string }) {
  const [status, setStatus] = useState('')
  return <div className="foundation-code">
    <div className="paper-cluster"><h3>{label}</h3><Button variant="outline" onClick={async () => {
      try { await navigator.clipboard.writeText(code); setStatus('Copied.') }
      catch { setStatus('Select the code below and copy it manually.') }
    }}>Copy code</Button><span role="status">{status}</span></div>
    <CodeBlock code={code} label={label} language={label.includes("CSS") ? "css" : label.includes("commands") ? "bash" : "tsx"} />
  </div>
}

export function GuideTable({ columns, rows, label }: { columns: readonly string[]; rows: readonly (readonly string[])[]; label: string }) {
  return <div className="foundation-reference" role="region" aria-label={label} tabIndex={0}>
    <p className="foundation-reference-hint">Scroll horizontally to see all columns.</p><table><caption>{label}</caption><thead><tr>{columns.map(c => <th key={c} scope="col">{c}</th>)}</tr></thead>
      <tbody>{rows.map((row, index) => <tr key={index}>{row.map((cell, i) => i === 0 ? <th key={i} scope="row">{cell}</th> : <td key={i}>{cell}</td>)}</tr>)}</tbody>
    </table>
  </div>
}

export function GovernanceGuide({ document }: { document: CatalogDocument }) {
  if (document.id === 'governance.contributing') return <section id="usage" className="document-section foundation-guide">
    <h2 className="paper-type-section">Make a reviewable change</h2>
    <ol>
      <li><strong>Check the existing contract.</strong> Read the component page, public export, CSS and tests in <code>frontend/packages/paper-ui</code>. Record the concrete user problem and whether an existing component can cover it.</li>
      <li><strong>Reproduce the behavior.</strong> For a bug, add a failing interaction or visual regression before fixing it. For a new API, describe defaults, controlled state, keyboard behavior and invalid combinations.</li>
      <li><strong>Update the package and catalogue together.</strong> Component sources live in <code>src/components</code>, tokens in <code>src/foundations</code>, recipes in <code>styles</code>, and document/fixture definitions in <code>src/catalog</code>. Export public APIs from <code>src/index.ts</code>.</li>
      <li><strong>Make the example runnable.</strong> Include imports, provider setup, initial state, callback behavior and all referenced variables. Cover empty, long, error, disabled and loading states when applicable. Keep fixture data deterministic.</li>
      <li><strong>Review the rendered result.</strong> Test keyboard entry/exit, focus, light/dark, comfortable/compact and narrow layout. Inspect long content and 200% zoom. Check the component in a real composition, not just alone.</li>
      <li><strong>Record evidence and consequences.</strong> Update review date and changelog with behavior/API impact, migration steps and unresolved decisions. Stable promotion requires human review; passing a content-presence test is insufficient.</li>
    </ol>
    <GuideCode label="Validation commands (repository root)" code={'cd frontend\npnpm --filter paper-ui build\npnpm --filter paper-ui typecheck\npnpm --filter paper-ui test\npnpm --filter paper-ui lint\npnpm --filter paper-ui consumer:ts49\npnpm --filter paper-styleguide typecheck\npnpm --filter paper-styleguide test\npnpm --filter paper-styleguide lint\npnpm --filter paper-styleguide build'} />
    <p>After editing package source, rebuild <code>paper-ui</code> before reviewing the styleguide: the app consumes the built package. Run <code>pnpm --filter paper-styleguide dev</code> from <code>frontend</code> to review locally.</p>
    <GuideTable label="Lifecycle is a promise" columns={['Status', 'Required meaning']} rows={[
      ['Experimental', 'Design or behavior still needs validation. Identify the open question; do not silently treat it as approved product behavior.'],
      ['Stable', 'Documented contract and complete examples, behavior evidence and visual review. Compatible changes preserve callers.'],
      ['Deprecated', 'Name the replacement, migration steps and agreed removal timing. Do not remove a used export without an exit path.'],
    ]} />
    <p><a href="https://github.com/tadoku/tadoku/tree/main/docs/wip/tadoku-paper">Decision log and design history</a> explain the accepted Bookplate direction. Keep application routing, fetching, persistence and contest rules in applications.</p>
  </section>
  if (document.id === 'governance.changelog') return <section id="usage" className="document-section foundation-guide">
    <h2 className="paper-type-section">Change history</h2>
    <p>The package version remains <code>0.1.0</code>. A documentation review date is not a release or deployment date. These entries describe repository work; verify the deployed commit before relying on a change in a live app.</p>
    <GuideTable label="Changes and consumer impact" columns={['When', 'Change', 'Consumer impact']} rows={[
      ['Current review · unreleased', 'Foundation usage examples, spacing scale and layout utilities; corrected accent rail perimeter alignment.', 'Existing composition remains valid. New utilities require the rebuilt package. Application migrations remain separate work.'],
      ['2026-08-09', 'Refined action focus, form borders and Flash rails; restored wordmark and heatmap; styleguide controls use Paper.', 'Visual changes to existing components. Catalogue coverage did not establish complete usage guidance or readiness to migrate every application.'],
      ['2026-08-08 · 0.1.0 baseline', 'Canonical catalogue, framework-neutral exports, semantic light/dark themes and density controls.', 'Private workspace package; React 18.2 and React Hook Form 7 peers. Built declarations target TypeScript 4.9 consumers.'],
    ]} />
    <h3>Before upgrading an application</h3><p>Read the affected component’s contract and migration notes, compare the application’s current commit to the intended version, then review its actual screens in both themes and densities. This page does not assert that Admin, Auth or webv2 has migrated.</p>
    <p><a href="https://github.com/tadoku/tadoku/commits/main/frontend/packages/paper-ui">Package commit history</a> and <a href="https://github.com/tadoku/tadoku/commits/main/frontend/apps/paper-styleguide">styleguide commit history</a> contain exact changes. Container digests and deployment approvals belong in release records, not inferred from the package version.</p>
  </section>
  if (document.id === 'pattern.logging' || document.id === 'experiment.logging-v2') return <section id="usage" className="document-section foundation-guide">
    <h2 className="paper-type-section">Compose the reading flow</h2>
    <GuideTable label="Information the reader must see" columns={['Moment', 'Required content', 'Application responsibility']} rows={[
      ['Entry', 'Work, language, amount/unit and reading date.', 'Validate values and preserve edits after failure.'],
      ['Review (when included)', 'Repeat the exact amount and work before saving.', 'Return to editing without losing values.'],
      ['Saved reading', 'Confirm whether saving succeeded and where to find the entry.', 'Only show success after the persistence request succeeds.'],
      ['Contest contribution', 'Explicit submitted / not submitted status; eligible contest if known.', 'Own eligibility and submission rules. Never infer submission merely from a saved log.'],
    ]} />
    <p>The preview is local and makes no network requests. Links and actions demonstrate navigation or in-page state; they do not save reading. In an application, replace them with routing and mutation handlers, handle pending/failure states, and keep successful entry data visible.</p>
    {document.id === 'experiment.logging-v2' ? <><h3>Still a product decision</h3><p>Whether review is mandatory, whether saving and contest submission are separate steps, and which contests are eligible need product validation. This prototype deliberately stops at review. Its confirmation is not proof of persistence or contest submission.</p></> : <><h3>Summary hierarchy</h3><p>Lead with the work and reading amount, then language/date and explicit contribution status. Keep notes privacy separate from contest visibility. Use a real link for “View log” and a button for an in-place action.</p></>}
  </section>
  return null
}
