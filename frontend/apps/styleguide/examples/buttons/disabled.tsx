export default function DisabledButtons() {
  return (
    <div className="h-stack spaced">
      <button className="btn primary" disabled>
        Primary
      </button>
      <button className="btn secondary" disabled>
        Secondary
      </button>
      <button className="btn" disabled>
        Tertiary (default)
      </button>
      <button className="btn danger" disabled>
        Danger
      </button>
      <button className="btn ghost" disabled>
        Ghost
      </button>
    </div>
  )
}
