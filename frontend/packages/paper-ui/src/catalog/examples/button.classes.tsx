export default function Example() {
  return (
    <div className="paper-cluster">
      <button type="button" className="paper-button paper-button--default">
        Save log
      </button>
      <a href="#logs" className="paper-button paper-button--outline">
        View logs
      </a>
      <button
        type="button"
        disabled
        aria-busy="true"
        className="paper-button paper-button--default paper-button--loading"
      >
        <span className="paper-button__spinner" aria-hidden="true" />
        <span className="paper-button__label">Saving log</span>
      </button>
    </div>
  )
}
