import { RankingRegistration, RawRankingRegistration } from './../interfaces'
import { Mapper, MappersWithOptional } from '../../interfaces'
import { Serializer } from '../../cache'
import { createSerializer, withOptional } from '../../transform'

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

export const rankingRegistrationMapper: MappersWithOptional<
  RawRankingRegistration,
  RankingRegistration
> = withOptional({
  toRaw: rankingRegistrationToRawMapper,
  fromRaw: rawToRankingRegistrationMapper,
})

export const rankingRegistrationSerializer: Serializer<RankingRegistration> = createSerializer(
  rankingRegistrationMapper,
)
