import { RankingRegistration, RawRankingRegistration } from './../interfaces'
import { Mapper } from '../../interfaces'
import { Serializer } from '../../cache'
import { withOptional } from '../../transform'

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

export const rankingRegistrationSerializer: Serializer<RankingRegistration> = {
  serialize: data => {
    const raw = rankingRegistrationToRawMapper(data)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return rawToRankingRegistrationMapper(raw)
  },
}

export const rankingRegistrationMapper = withOptional({
  toRaw: rankingRegistrationToRawMapper,
  fromRaw: rawToRankingRegistrationMapper,
})
