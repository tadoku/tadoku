import { cluster, Loading, stack } from 'paper-ui'

export default function Example() {
  return (
    <div className={stack()}>
      <p>Loading reading entries. Your saved entries will appear here.</p>
      <div className={cluster()}>
        <span className={stack()}>
          <Loading
            label="Loading reading entries, small example"
            size="small"
          />
          <span>Small</span>
        </span>
        <span className={stack()}>
          <Loading label="Loading reading entries, default example" />
          <span>Default</span>
        </span>
        <span className={stack()}>
          <Loading
            label="Loading reading entries, large example"
            size="large"
          />
          <span>Large</span>
        </span>
      </div>
    </div>
  )
}
