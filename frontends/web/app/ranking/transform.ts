import { AggregatedContestLogsByDayEntry, ContestLog } from './interfaces'

type AggregatedByDaysResult = {
  [key: string]: AggregatedContestLogsByDayEntry[]
}

export const aggregateContestLogsByDays = (
  logs: ContestLog[],
): AggregatedByDaysResult => {
  const aggregated: {
    [key: string]: { [key: string]: AggregatedContestLogsByDayEntry }
  } = {}

  logs.forEach(log => {
    if (!Object.keys(aggregated).includes(log.languageCode)) {
      aggregated[log.languageCode] = {}
    }

    const date = prettyDate(log.date)

    if (!Object.keys(aggregated[log.languageCode]).includes(date)) {
      aggregated[log.languageCode][date] = { x: new Date(date), y: 0 }
    }

    aggregated[log.languageCode][date].y +=
      Math.round(log.adjustedAmount * 10) / 10
  })

  const result: AggregatedByDaysResult = {}

  Object.keys(aggregated).forEach(languageCode => {
    result[languageCode] = Object.values(aggregated[languageCode])
  })

  return result
}

export const prettyDate = (date: Date): string =>
  `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
