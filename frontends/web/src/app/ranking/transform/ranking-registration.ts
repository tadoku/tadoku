import { RankingRegistration, RawRankingRegistration } from './../interfaces'
import { Mapper, Mappers } from '../../interfaces'
import { Serializer } from '../../cache'
import { createSerializer, createMappers } from '../../transform'

const rawToRankingRegistrationMapper: Mapper<
  RawRankingRegistration,
  RankingRegistration
> = raw => ({
  contestId: raw.contest_id,
  languages: raw.languages,
  start: new Date(raw.start),
  end: new Date(raw.end),
})

const rankingRegistrationToRawMapper: Mapper<
  RankingRegistration,
  RawRankingRegistration
> = registration => ({
  contest_id: registration.contestId,
  languages: registration.languages,
  start: registration.start.toISOString(),
  end: registration.end.toISOString(),
})

export const rankingRegistrationMapper: Mappers<
  RawRankingRegistration,
  RankingRegistration
> = createMappers({
  toRaw: rankingRegistrationToRawMapper,
  fromRaw: rawToRankingRegistrationMapper,
})

export const rankingRegistrationSerializer: Serializer<RankingRegistration> =
  createSerializer(rankingRegistrationMapper)
