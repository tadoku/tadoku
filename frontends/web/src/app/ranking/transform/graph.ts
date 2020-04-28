import {
  AggregatedContestLogsByDayEntry,
  ContestLog,
  Ranking,
  RankingRegistrationOverview,
  RankingWithRank,
} from '../interfaces'
import { Contest } from '../../contest/interfaces'
import {
  languageNameByCode,
  mediumDescriptionById,
  GlobalLanguage,
} from '../database'
import { graphColor } from '../../ui/components/Graphs'

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

interface AggregatedByDaysResult {
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
    initializedSeries[prettyDate(date)] = {
      x: date,
      y: 0,
      language: '',
      size: 2,
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

  const result: AggregatedByDaysResult = { aggregated: {}, legend }

  Object.keys(aggregated).forEach(languageCode => {
    result.aggregated[languageCode] = Object.values(aggregated[languageCode])
  })

  return result
}

interface AggregatedByMediumResult {
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

export const aggregateContestLogsByMedium = (
  logs: ContestLog[],
): AggregatedByMediumResult => {
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

export const amountToPages = (amount: number) => Math.round(amount * 10) / 10

export const pagesLabel = (languageCode: string) => {
  if (languageCode == GlobalLanguage.code) {
    return 'Overall score'
  }

  return `Score for ${languageNameByCode(languageCode)}`
}

export const calculateLeaderboard = (
  rankings: Ranking[],
): RankingWithRank[] => {
  const initialRankingsByScore: { [key: number]: Ranking[] } = {}

  const rankingsByScore = rankings.reduce((scores, ranking) => {
    if (!scores[ranking.amount]) {
      scores[ranking.amount] = []
    }

    scores[ranking.amount].push(ranking)

    return scores
  }, initialRankingsByScore)

  const sortedScores = Object.keys(rankingsByScore)
    .map(n => Number(n))
    .sort((a, b) => b - a)

  const initialResult: {
    currentRank: number
    rankings: RankingWithRank[]
  } = {
    currentRank: 1,
    rankings: [],
  }

  const result = sortedScores.reduce((result, score) => {
    const rankings = rankingsByScore[score]

    const newRankings = rankings.map(ranking => ({
      data: ranking,
      tied: rankings.length > 1,
      rank: result.currentRank,
    }))

    result.rankings.push(...newRankings)
    result.currentRank += rankings.length

    return result
  }, initialResult)

  return result.rankings
}

export const amountToString = (amount: number): string => {
  switch (amount) {
    case 0:
      return 'No pages'
    case 1:
      return '1 page'
    default:
      return `${amountToPages(amount)} pages`
  }
}
