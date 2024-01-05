import { CodeBlock, Preview, Title } from '@components/example'

export default function Page() {
  return (
    <>
      <h1 className="title mb-8">Logging flow</h1>

      <Title>Logs overview</Title>
      <LogList />
    </>
  )
}

/**
 * Stuff to list: activity, date, language, description, amount/score toggle, duration
 *
 */

interface LogSummary {
  date: string
  language: string
  activity: string
  logs: Log[]
}

interface Log {
  duration?: number
  score: number
  amount: number
  unit: string
  description?: string
  tags?: string[]
}

function LogList() {
  const summaries: LogSummary[] = [
    {
      date: 'Oct 1, 2023',
      language: 'Japanese',
      activity: 'Reading',
      logs: [
        {
          duration: 35,
          amount: 40,
          score: 8,
          unit: 'comic pages',
          description: 'Bleach ch 60-61',
          tags: ['manga', 'fiction'],
        },
        { amount: 20, score: 4, unit: 'comic pages' },
        {
          duration: 80,
          amount: 40,
          score: 40,
          unit: 'pages',
          description: 'Sword Art Online v15 ch7',
          tags: ['lightnovel', 'fiction'],
        },
      ],
    },
    {
      date: 'Oct 1, 2023',
      language: 'Chinese',
      activity: 'Listening',
      logs: [
        {
          duration: 30,
          amount: 30,
          score: 15,
          unit: 'minutes',
          tags: ['radio'],
        },
      ],
    },
    {
      date: 'Sep 30, 2023',
      language: 'Chinese',
      activity: 'Reading',
      logs: [
        { amount: 40, score: 40, unit: 'pages' },
        { amount: 3, score: 3, unit: 'pages' },
      ],
    },
  ]

  const totalDurationFor = (summary: LogSummary) =>
    summary.logs
      .map(it => it.duration ?? 0)
      .reduce((total, current) => total + current, 0)

  const totalScoreFor = (summary: LogSummary) =>
    summary.logs
      .map(it => it.score)
      .reduce((total, current) => total + current, 0)

  const formatDuration = (minutes: number | undefined) => {
    if (!minutes || minutes == 0) {
      return 'N/A'
    }

    return `${Math.floor(minutes / 60)}:${minutes % 60}`
  }

  return (
    <div className="table-container shadow-transparent w-auto">
      <table className="default shadow-transparent">
        <thead>
          <tr>
            <th className="default">Log</th>
            <th className="default w-32 !text-right">Score</th>
            <th className="default w-24 !text-right">Duration</th>
          </tr>
        </thead>
        <tbody>
          {summaries.map(summary => (
            <>
              <tr>
                <td className="default flex items-center flex-wrap space-x-4 text-sm">
                  <div className="flex items-center space-x-2">
                    <div
                      className={`w-[6px] h-[6px] rounded-lg ${
                        summary.language == 'Japanese'
                          ? 'bg-red-700'
                          : 'bg-slate-600'
                      }`}
                    ></div>
                    <span
                      className={`${
                        summary.language == 'Japanese'
                          ? 'text-red-700'
                          : 'text-slate-600'
                      }`}
                    >
                      {summary.language}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div
                      className={`w-[6px] h-[6px] rounded-lg ${
                        summary.activity == 'Reading'
                          ? 'bg-emerald-700'
                          : 'bg-amber-600'
                      }`}
                    ></div>
                    <span
                      className={
                        summary.activity == 'Reading'
                          ? 'text-emerald-700'
                          : 'text-amber-600'
                      }
                    >
                      {summary.activity}
                    </span>
                  </div>
                  <span>{summary.date}</span>
                </td>
                <td className="default font-bold text-right">
                  {totalScoreFor(summary)}
                </td>
                <td className="default font-bold text-right">
                  {formatDuration(totalDurationFor(summary))}
                </td>
              </tr>
              {summary.logs.map((log, idx) => (
                <tr
                  className={`text-sm bg-slate-500/5 border-slate-500/5 ${
                    idx === summary.logs.length - 1 ? 'border-b-2' : 'border-b'
                  }`}
                >
                  <td className="default !pl-7 flex">
                    <span className="grow flex items-center">
                      {log.description == undefined ? (
                        ''
                      ) : (
                        <>
                          <strong>{log.description}</strong>,{' '}
                        </>
                      )}
                      {log.amount} {log.unit}
                    </span>
                    {(log.tags ?? []).length > 0 ? (
                      <span className="opacity-50 items-center flex">
                        {log.tags!!.map(it => `#${it}`).join(' ')}
                      </span>
                    ) : null}
                  </td>
                  <td className="default text-right">{log.score}</td>
                  <td className="default text-right">
                    {formatDuration(log.duration)}
                  </td>
                </tr>
              ))}
            </>
          ))}
        </tbody>
      </table>
    </div>
  )
}
