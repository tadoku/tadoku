import { buttonClassName, cluster } from 'paper-ui'

export default function Example() {
  return (
    <div className={cluster()}>
      <button type="button" className={buttonClassName()}>
        Save log
      </button>
      <a href="#logs" className={buttonClassName({ variant: 'outline' })}>
        View logs
      </a>
      <button type="button" disabled aria-busy="true" className={buttonClassName({ loading: true })}>
        Saving log
      </button>
    </div>
  )
}
