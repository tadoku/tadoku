import { useStore } from '@tanstack/react-form'
import { usePaperForm } from 'paper-ui'
import type { CatalogDocument } from 'paper-ui/catalog'
import { PrinciplesSpecimen } from './foundation-specimens/PrinciplesSpecimen'
import { ColorSpecimen } from './foundation-specimens/ColorSpecimen'
import { TypographySpecimen } from './foundation-specimens/TypographySpecimen'
import { SpacingSpecimen } from './foundation-specimens/SpacingSpecimen'
import { LayoutSpecimen } from './foundation-specimens/LayoutSpecimen'
import { ShapeSpecimen } from './foundation-specimens/ShapeSpecimen'
import { ElevationSpecimen } from './foundation-specimens/ElevationSpecimen'
import { IconographySpecimen } from './foundation-specimens/IconographySpecimen'
import { MotionSpecimen } from './foundation-specimens/MotionSpecimen'
import { BrandSpecimen } from './foundation-specimens/BrandSpecimen'
import { PreviewSettings } from './foundation-specimens/preview-settings'

const specimens: Readonly<Record<string, () => JSX.Element>> = {
  'foundation.principles': PrinciplesSpecimen, 'foundation.color': ColorSpecimen,
  'foundation.typography': TypographySpecimen, 'foundation.spacing-and-density': SpacingSpecimen,
  'foundation.layout': LayoutSpecimen, 'foundation.shape-and-borders': ShapeSpecimen,
  'foundation.elevation': ElevationSpecimen, 'foundation.iconography': IconographySpecimen,
  'foundation.motion': MotionSpecimen, 'foundation.brand': BrandSpecimen,
}
export function FoundationSpecimen({ document }: { document: CatalogDocument }) {
  const Specimen = specimens[document.id]
  const methods = usePaperForm({ defaultValues: { theme: 'light', density: 'comfortable' } })
  const theme = useStore(methods.store, state => state.values['theme'])
  const density = useStore(methods.store, state => state.values['density'])
  return <section className="foundation-guide" aria-label={`${document.name} specimen`} data-foundation-specimen={document.id}>
    <methods.AppForm><div className="foundation-guide__grid" role="group" aria-label="Foundation preview settings">
      <methods.AppField name="theme">{field => <field.Select label="Example theme" options={[{value: 'light', label: 'Light'}, {value: 'dark', label: 'Dark'}]} />}</methods.AppField>
      <methods.AppField name="density">{field => <field.Select label="Example density" options={[{value: 'comfortable', label: 'Comfortable'}, {value: 'compact', label: 'Compact'}]} />}</methods.AppField>
    </div></methods.AppForm>
    <p className="foundation-guide__settings-note">These controls affect the visual examples below. See <a href="/setup">Setup</a> for imports and application settings.</p>
    <PreviewSettings.Provider value={{ theme, density }}>
      {Specimen ? <Specimen/> : null}
    </PreviewSettings.Provider>
  </section>
}
