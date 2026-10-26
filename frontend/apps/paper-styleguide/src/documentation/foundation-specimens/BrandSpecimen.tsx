import BrandLinkExample from '../foundation-examples/BrandLinkExample'
import BrandLinkExampleSource from '../foundation-examples/BrandLinkExample.tsx?raw'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import cutMeterReversedUrl from 'paper-ui/assets/brand/cut-meter-reversed.svg?no-inline'
import wordmarkUrl from 'paper-ui/assets/brand/wordmark-accent.svg?no-inline'
import wordmarkReversedUrl from 'paper-ui/assets/brand/wordmark-reversed.svg?no-inline'
import { GuideTable } from '../FoundationGuide'
import { FoundationCanvas, FoundationExample } from './FoundationPreview'

export function BrandSpecimen() {
  return <><FoundationCanvas><div className="foundation-guide__grid foundation-brand-guide">{(['light', 'dark'] as const).map(theme => <figure key={theme} data-theme={theme} className="foundation-guide__sample paper-stack"><img src={theme === 'light' ? wordmarkUrl : wordmarkReversedUrl} width={158} height={29} alt="Tadoku"/><img src={theme === 'light' ? cutMeterUrl : cutMeterReversedUrl} width={64} height={64} alt=""/><figcaption>{theme === 'light' ? 'Original wordmark on light paper' : 'Reversed on dark paper'}</figcaption></figure>)}</div></FoundationCanvas>
    <GuideTable label="Choose the packaged artwork" columns={['Asset', 'Use']} rows={[
      ['wordmark-accent.svg / wordmark-reversed.svg', 'Primary identity where the full Tadoku name fits; preserve 158:29 proportions.'], ['cut-meter.svg / cut-meter-reversed.svg', 'Compact identity when the name is already nearby or an accessible label supplies it.'], ['wordmark.svg / cut-meter.svg', 'Monochrome artwork for contexts that require a single ink. The normal wordmark keeps its original violet K.'], ['cut-meter-wordmark.svg', 'Packaged combined lockup; do not rebuild from individually positioned bars.'], ['cut-meter.svg + wordmark-accent.svg; reversed pair on dark', 'Application header: the whole mark before the wordmark, its side 28% of the wordmark width with a 5.6% gap. The mark is decorative because the wordmark names Tadoku.'], ['favicon.svg; favicon-16.svg, -32.svg, -48.svg', 'Small browser identity with the approved small-size geometry.'], ['apple-touch-icon.svg', 'Opaque 180px source; rasterize when the target platform requires PNG.'],
    ]}/>
    <FoundationExample Example={BrandLinkExample} label="Brand link (Vite asset imports)" code={BrandLinkExampleSource} />
    <p>The asset import above assumes Vite’s asset handling and <code>vite/client</code> types. Other bundlers should copy the packaged file to a public asset URL. External SVG images do not inherit the page’s currentColor; choose the reversed asset for dark surfaces instead of applying a CSS filter.</p><p>Keep artwork unstretched and preserve its view box. Supply one accessible name per link: if the link already says “Tadoku home,” the image is decorative. Formal minimum size and clear-space specifications remain a brand decision; the examples show current package dimensions rather than inventing a new standard.</p>
  </>
}

