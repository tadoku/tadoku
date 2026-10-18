import { Language } from '@app/immersion/api'
import { Score } from '@app/immersion/scoring-api'
import { formatScore } from '@app/common/format'
import { DateTime } from 'luxon'

interface Props {
  list: Score[]
  languages: Language[]
  year?: number
}

export const ScoreList = ({ list, languages, year }: Props) => {
  const total = list.map(it => it.score).reduce((prev, cur) => prev + cur, 0)
  const today = DateTime.utc()
  let days = 0
  if (year !== undefined && year <= today.year) {
    days = year === today.year ? today.ordinal : DateTime.utc(year).daysInYear
  }

  const scores = new Map<string, number>()
  for (const { language_code, score } of list) {
    scores.set(language_code, score)
  }

  return (
    <div className="w-full grid gap-4">
      <div className="card narrow">
        <h3 className="subtitle mb-2">Overall score</h3>
        <span className="text-4xl font-bold">{formatScore(total)}</span>
        {days > 0 && total > 0 ? (
          <p className="mt-2 text-sm text-slate-500">
            <span className="font-bold">{formatScore(total / days)}</span> per
            day · average over{' '}
            {year === today.year
              ? `the ${days} ${days === 1 ? 'day' : 'days'} of ${year} so far`
              : `all ${days} days of ${year}`}
          </p>
        ) : null}
      </div>
      {languages.map(({ code, name }) => (
        <div className="card narrow" key={code}>
          <h3 className="subtitle mb-2">{name}</h3>
          <span className="text-4xl font-bold">
            {formatScore(scores.get(code) ?? 0)}
          </span>
        </div>
      ))}
    </div>
  )
}
