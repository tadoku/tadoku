export default function ReadingTotals() {
  return (
    <div
      role="region"
      aria-label="Reading totals"
      tabIndex={0}
      style={{ maxWidth: '100%', overflowX: 'auto' }}
    >
      <table style={{ minWidth: '34rem' }}>
        <caption>Reading totals</caption>
        <thead>
          <tr>
            <th scope="col">Language</th>
            <th scope="col">Pages</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <th scope="row">Japanese</th>
            <td>96</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}
