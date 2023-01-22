import { ContestView } from '@app/immersion/api'

interface Props {
  contest: ContestView
}

export const ContestConfiguration = ({ contest }: Props) => (
  <div className="card text-sm">
    <div className="-m-7 pt-4 px-4 text-sm">
      <h3 className="subtitle text-sm mb-2">Configuration</h3>
      <div className="flex mb-2">
        <span className="w-1/2">Languages</span>
        <div>
          {!contest.allowed_languages ||
          contest.allowed_languages.length === 0 ? (
            <strong>No restrictions</strong>
          ) : (
            <ul>
              {contest.allowed_languages.map(it => (
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
            {contest.allowed_activities.map(it => (
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