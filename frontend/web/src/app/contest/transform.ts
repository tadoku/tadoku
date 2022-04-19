import {
  Contest,
  RawContestStats,
  ContestStats,
  RawContest,
} from './interfaces'
import { Mapper, Mappers } from '../interfaces'
import { Serializer } from '../cache'
import {
  createSerializer,
  createMappers,
  createCollectionSerializer,
} from '../transform'

const rawToContestMapper: Mapper<RawContest, Contest> = raw => ({
  id: raw.id,
  description: raw.description,
  start: new Date(raw.start),
  end: new Date(raw.end),
  open: raw.open,
})

const contestToRawMapper: Mapper<Contest, RawContest> = contest => ({
  id: contest.id,
  description: contest.description,
  start: contest.start.toISOString(),
  end: contest.end.toISOString(),
  open: contest.open,
})

export const contestMapper: Mappers<RawContest, Contest> = createMappers({
  fromRaw: rawToContestMapper,
  toRaw: contestToRawMapper,
})

export const contestSerializer: Serializer<Contest> =
  createSerializer(contestMapper)

export const contestCollectionSerializer: Serializer<Contest[]> =
  createCollectionSerializer(contestMapper)

const rawToContestStatsMapper: Mapper<RawContestStats, ContestStats> = raw => ({
  totalAmount: raw.total_amount,
  participants: raw.participants,
  byLanguage: raw.by_language.map(({ count, language_code }) => ({
    count,
    languageCode: language_code,
  })),
})

const contestStatsToRawMapper: Mapper<ContestStats, RawContestStats> =
  stats => ({
    total_amount: stats.totalAmount,
    participants: stats.participants,
    by_language: stats.byLanguage.map(({ count, languageCode }) => ({
      count,
      language_code: languageCode,
    })),
  })

export const contestStatsMapper: Mappers<RawContestStats, ContestStats> =
  createMappers({
    toRaw: contestStatsToRawMapper,
    fromRaw: rawToContestStatsMapper,
  })
