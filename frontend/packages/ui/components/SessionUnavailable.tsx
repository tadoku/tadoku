export function SessionUnavailable() {
  return (
    <div role="alert" className="p-8 space-y-4">
      <h1 className="text-xl font-bold">Session temporarily unavailable</h1>
      <p>We couldn’t check your session. Please try again.</p>
      <button className="btn primary" onClick={() => window.location.reload()}>
        Try again
      </button>
    </div>
  )
}
