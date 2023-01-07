import { ContestView } from '@app/contests/api'

interface Props {
  contest: ContestView
}

export const ContestConfiguration = ({ contest }: Props) => (
  <div className="card text-sm">
    <div className="-m-7 pt-4 px-4 text-sm">
      <h3 className="subtitle text-sm mb-2">Contest Configuration</h3>
      <div className="flex mb-2">
        <span className="w-1/2">Languages</span>
        <div>
          {!contest.allowedLanguages ||
          contest.allowedLanguages.length === 0 ? (
            <strong>No restrictions</strong>
          ) : (
            <ul>
              {contest.allowedLanguages.map(it => (
                <li key={it.code}>
                  <strong>{it.name}</strong>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
      <div className="flex mb-4">
        <span className="w-1/2">Activities</span>
        <div>
          <ul>
            {contest.allowedActivities.map(it => (
              <li className="font-bold" key={it.id}>
                {it.name}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  </div>
)
