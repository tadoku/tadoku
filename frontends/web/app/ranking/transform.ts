import {
  AggregatedContestLogsByDayEntry,
  ContestLog,
  Ranking,
  RankingRegistrationOverview,
} from './interfaces'
import { Contest } from './../contest/interfaces'
import { languageNameByCode } from './database'

type AggregatedByDaysResult = {
  aggregated: {
    [languageCode: string]: AggregatedContestLogsByDayEntry[]
  }
  legend: {
    title: string
  }[]
}

export const aggregateContestLogsByDays = (
  logs: ContestLog[],
  contest: Contest,
): AggregatedByDaysResult => {
  const aggregated: {
    [languageCode: string]: { [date: string]: AggregatedContestLogsByDayEntry }
  } = {}

  const languages: string[] = []
  const legend: {
    title: string
  }[] = []

  logs.forEach(log => {
    if (!languages.includes(log.languageCode)) {
      languages.push(log.languageCode)
      legend.push({
        title: languageNameByCode(log.languageCode),
      })
    }
  })

  const initializedSeries: {
    [date: string]: AggregatedContestLogsByDayEntry
  } = {}

  getDates(contest.start, contest.end).forEach(date => {
    initializedSeries[prettyDate(date)] = { x: date, y: 0 }
  })

  languages.forEach(language => {
    const series: typeof initializedSeries = {}

    Object.keys(initializedSeries).forEach(date => {
      series[date] = { ...initializedSeries[date] }
    })

    aggregated[language] = series
  })

  logs.forEach(log => {
    const date = prettyDate(log.date)

    if (aggregated[log.languageCode][date]) {
      aggregated[log.languageCode][date].y +=
        Math.round(log.adjustedAmount * 10) / 10
    }
  })

  const result: AggregatedByDaysResult = { aggregated: {}, legend }

  Object.keys(aggregated).forEach(languageCode => {
    result.aggregated[languageCode] = Object.values(aggregated[languageCode])
  })

  return result
}

export const prettyDate = (date: Date): string =>
  `${date.getUTCFullYear()}-${date.getUTCMonth() + 1}-${date.getUTCDate()}`

const getDates = (startDate: Date, endDate: Date) => {
  const dates = []

  let currentDate = new Date(
    Date.UTC(
      startDate.getUTCFullYear(),
      startDate.getUTCMonth(),
      startDate.getUTCDate(),
    ),
  )

  while (currentDate <= endDate) {
    dates.push(currentDate)

    currentDate = new Date(
      Date.UTC(
        currentDate.getUTCFullYear(),
        currentDate.getUTCMonth(),
        currentDate.getUTCDate() + 1,
      ),
    )
  }

  return dates
}

export const rankingsToRegistrationOverview = (
  rankings: Ranking[],
): RankingRegistrationOverview | undefined => {
  if (rankings.length === 0) {
    return undefined
  }

  const registrations = rankings
    .map(r => ({
      languageCode: r.languageCode,
      amount: r.amount,
    }))
    .reduce(
      (acc, element) => {
        if (element.languageCode === 'GLO') {
          return [element, ...acc]
        }
        return [...acc, element]
      },
      [] as { languageCode: string; amount: number }[],
    )

  return {
    contestId: rankings[0].contestId,
    userId: rankings[0].userId,
    userDisplayName: rankings[0].userDisplayName,
    registrations,
  }
}

export const amountToPages = (amount: number) => Math.round(amount * 10) / 10

export const pagesLabel = (languageCode: string) =>
  `pages in ${languageNameByCode(languageCode)}`
