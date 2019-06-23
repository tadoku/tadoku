import {
  AggregatedContestLogsByDayEntry,
  ContestLog,
  Ranking,
  RankingRegistrationOverview,
  RawRanking,
  RawContestLog,
  RankingRegistration,
  RawRankingRegistration,
  RankingWithRank,
} from './interfaces'
import { Contest } from './../contest/interfaces'
import { languageNameByCode, mediumDescriptionById } from './database'
import { Mapper } from '../interfaces'
import { Serializer } from '../cache'

export const RawToRankingMapper: Mapper<RawRanking, Ranking> = raw => ({
  contestId: raw.contest_id,
  userId: raw.user_id,
  userDisplayName: raw.user_display_name,
  languageCode: raw.language_code,
  amount: raw.amount,
})

export const RankingToRawMapper: Mapper<Ranking, RawRanking> = ranking => ({
  contest_id: ranking.contestId,
  user_id: ranking.userId,
  user_display_name: ranking.userDisplayName,
  language_code: ranking.languageCode,
  amount: ranking.amount,
})

export const RankingsSerializer: Serializer<Ranking[]> = {
  serialize: data => {
    const raw = data.map(RankingToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(RawToRankingMapper)
  },
}

export const RawToContestLogMapper: Mapper<
  RawContestLog,
  ContestLog
> = raw => ({
  id: raw.id,
  contestId: raw.contest_id,
  userId: raw.user_id,
  languageCode: raw.language_code,
  mediumId: raw.medium_id,
  amount: raw.amount,
  adjustedAmount: raw.adjusted_amount,
  date: new Date(raw.date),
})

export const ContestLogToRawMapper: Mapper<
  ContestLog,
  RawContestLog
> = contestLog => ({
  id: contestLog.id,
  contest_id: contestLog.contestId,
  user_id: contestLog.userId,
  language_code: contestLog.languageCode,
  medium_id: contestLog.mediumId,
  amount: contestLog.amount,
  adjusted_amount: contestLog.adjustedAmount,
  date: contestLog.date.toISOString(),
})

export const ContestLogsSerializer: Serializer<ContestLog[]> = {
  serialize: data => {
    const raw = data.map(ContestLogToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(RawToContestLogMapper)
  },
}

export const RawToRankingRegistrationMapper: Mapper<
  RawRankingRegistration,
  RankingRegistration
> = raw => ({
  contestId: raw.contest_id,
  languages: raw.languages,
  start: new Date(raw.start),
  end: new Date(raw.end),
})

export const RankingRegistrationToRawMapper: Mapper<
  RankingRegistration,
  RawRankingRegistration
> = registration => ({
  contest_id: registration.contestId,
  languages: registration.languages,
  start: registration.start.toISOString(),
  end: registration.end.toISOString(),
})

export const RankingRegistrationSerializer: Serializer<RankingRegistration> = {
  serialize: data => {
    const raw = RankingRegistrationToRawMapper(data)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return RawToRankingRegistrationMapper(raw)
  },
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
    initializedSeries[prettyDate(date)] = { x: date, y: 0, language: '' }
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
  }[]
  totalAmount: number
  legend: {
    title: string
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

  const forChart = Object.keys(aggregated)
    .map(k => Number(k))
    .map(k => ({
      amount: aggregated[k],
      medium: mediumDescriptionById(k),
    }))

  const legend = Object.values(forChart)
    .sort((a, b) => a.amount - b.amount)
    .map(mediumStats => ({
      title: mediumStats.medium,
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
