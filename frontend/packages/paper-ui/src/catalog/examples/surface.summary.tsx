import { stack, Surface } from 'paper-ui'

export default function Example() {
  return (
    <div className={stack({ className: "max-w-form" })}>
      {(['flat', 'floating', 'showcase'] as const).map(elevation => (
        <Surface
          key={elevation}
          as="article"
          elevation={elevation}
          className={stack()}
        >
          <h3 className="m-0">
            {elevation === 'flat'
              ? 'Flat'
              : elevation === 'floating'
              ? 'Floating'
              : 'Showcase'}{' '}
            summary
          </h3>
          <p className="m-0">
            August Japanese: 1,240 pages across 18 entries.
          </p>
        </Surface>
      ))}
      <Surface as="article" accent className={stack()}>
        <h3 className="m-0">Flat with accent rail</h3>
        <p className="m-0">
          The same reading summary with an independent accent.
        </p>
      </Surface>
    </div>
  )
}
