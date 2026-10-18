export default function ScrollingTable() {
  return (
    <div className="max-w-xs">
      <h3 id="scrolling-table-title" className="font-semibold mb-2">
        Activity summary
      </h3>
      <div
        className="table-container"
        role="region"
        tabIndex={0}
        aria-labelledby="scrolling-table-title"
      >
        <table className="default">
          <thead>
            <tr>
              <th className="default">Language</th>
              <th className="default">Reading</th>
              <th className="default">Writing</th>
              <th className="default">Listening</th>
              <th className="default">Speaking</th>
              <th className="default">Study</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="default">Japanese</td>
              <td className="default">120</td>
              <td className="default">15</td>
              <td className="default">60</td>
              <td className="default">30</td>
              <td className="default">45</td>
            </tr>
            <tr>
              <td className="default">Spanish</td>
              <td className="default">80</td>
              <td className="default">10</td>
              <td className="default">90</td>
              <td className="default">20</td>
              <td className="default">30</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}
