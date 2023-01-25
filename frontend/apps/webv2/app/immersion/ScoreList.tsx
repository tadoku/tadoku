import { Language, Score } from '@app/immersion/api'
import { formatScore } from '@app/common/format'

interface Props {
  list: Score[]
  languages: Language[]
}

export const ScoreList = ({ list, languages }: Props) => {
  const total = list.map(it => it.score).reduce((prev, cur) => prev + cur, 0)

  const scores = new Map<string, number>()
  for (const { language_code, score } of list) {
    scores.set(language_code, score)
  }

  return (
    <div className="w-full grid gap-4">
      <div className="card narrow">
        <h3 className="subtitle mb-2">Overall score</h3>
        <span className="text-4xl font-bold">{formatScore(total)}</span>
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
