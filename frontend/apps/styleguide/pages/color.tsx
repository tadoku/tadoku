import { Logo, LogoInverted } from '@components/branding'
import { CodeBlock, Preview, Separator, Title } from '@components/example'

const colors = [
  ['primary', 'Primary color used in Tadoku'],
  ['secondary', 'Secondary color used in Tadoku'],
  ['slate-500', 'Used for neutral text that should not stand out'],
  ['red-600', 'Used for dangerous operations'],
  ['lime-700', 'Used to indicate success'],
  ['neutral-100', 'Used for light text'],
  ['neutral-900', 'Used for dark text'],
]

export default function branding() {
  return (
    <>
      <h1 className="title mb-8">Color</h1>
      <table className="w-full">
        <thead>
          <tr>
            <td className="subtitle px-4 py-2">Name</td>
            <td className="subtitle px-4 py-2">Description</td>
          </tr>
        </thead>
        <tbody>
          {colors.map(([color, description]) => (
            <tr key={color}>
              <td className="flex items-center px-4 py-2">
                <div className={`bg-${color} h-10 w-10 mr-4`}></div>
                <div className="align-middle font-mono">{color}</div>
              </td>
              <td className="px-4 py-2">{description}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  )
}
