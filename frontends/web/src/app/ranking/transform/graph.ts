import { ContestLog, Ranking, RankingRegistrationOverview } from '../interfaces'
import { Contest } from '../../contest/interfaces'
import { languageNameByCode, mediumDescriptionById } from '../database'
import { graphColor } from '../../ui/components/Graphs'

// Utils
const prettyDate = (date: Date): string =>
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

  while (currentDate < endDate) {
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

// Graph aggregators

interface AggregatedReadingActivity {
  aggregated: {
    [languageCode: string]: AggregatedReadingActivityEntry[]
  }
  legend: {
    title: string
  }[]
}

export interface AggregatedReadingActivityEntry {
  x: string // date in iso string for x axis
  y: number // page count for y axis
  language: string
}

export const aggregateReadingActivity = (
  logs: ContestLog[],
  contest: Contest,
): AggregatedReadingActivity => {
  const aggregated: {
    [languageCode: string]: { [date: string]: AggregatedReadingActivityEntry }
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
    [date: string]: AggregatedReadingActivityEntry
  } = {}

  getDates(contest.start, contest.end).forEach(date => {
    initializedSeries[prettyDate(date)] = {
      x: date.toISOString(),
      y: 0,
      language: '',
    }
  })

  languages.forEach(language => {
    const series: typeof initializedSeries = {}

    Object.keys(initializedSeries).forEach(date => {
      series[date] = {
        ...initializedSeries[date],
        language: languageNameByCode(language),
      }
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

  const result: AggregatedReadingActivity = { aggregated: {}, legend }

  Object.keys(aggregated).forEach(languageCode => {
    result.aggregated[languageCode] = Object.values(aggregated[languageCode])
  })

  return result
}

interface AggregatedMediaDistribution {
  aggregated: {
    amount: number
    medium: string
    color: string
  }[]
  totalAmount: number
  legend: {
    title: string
    strokeWidth: number
    color: string
    amount: number
  }[]
}

export const aggregateMediaDistribution = (
  logs: ContestLog[],
): AggregatedMediaDistribution => {
  const aggregated: {
    [mediumId: number]: number
  } = {}

  let total = 0
  logs.forEach(log => {
    if (!Object.keys(aggregated).includes(log.mediumId.toString())) {
      aggregated[log.mediumId] = 0
    }

    aggregated[log.mediumId] += log.adjustedAmount
    total += log.adjustedAmount
  })

  let forChart = Object.keys(aggregated)
    .map(k => Number(k))
    .map((k, i) => ({
      amount: aggregated[k],
      medium: mediumDescriptionById(k),
      color: graphColor(i),
    }))

  forChart.sort((a, b) => b.amount - a.amount)

  let legend = Object.values(forChart).map(mediumStats => ({
    title: mediumStats.medium,
    color: mediumStats.color,
    strokeWidth: 10,
    amount: mediumStats.amount,
  }))

  return { aggregated: forChart, legend, totalAmount: total }
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
    .reduce((acc, element) => {
      if (element.languageCode === 'GLO') {
        return [element, ...acc]
      }
      return [...acc, element]
    }, [] as { languageCode: string; amount: number }[])

  return {
    contestId: rankings[0].contestId,
    userId: rankings[0].userId,
    userDisplayName: rankings[0].userDisplayName,
    registrations,
  }
}
