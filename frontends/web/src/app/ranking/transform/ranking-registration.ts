import { RankingRegistration, RawRankingRegistration } from './../interfaces'
import { Mapper } from '../../interfaces'
import { Serializer } from '../../cache'

const RawToRankingRegistrationMapper: Mapper<
  RawRankingRegistration,
  RankingRegistration
> = raw => ({
  contestId: raw.contest_id,
  languages: raw.languages,
  start: new Date(raw.start),
  end: new Date(raw.end),
})

const RankingRegistrationToRawMapper: Mapper<
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

export const RankingRegistrationMapper = {
  toRaw: RankingRegistrationToRawMapper,
  fromRaw: RawToRankingRegistrationMapper,
  optional: {
    toRaw: (
      registration: RankingRegistration | undefined,
    ): RawRankingRegistration | undefined =>
      registration ? RankingRegistrationToRawMapper(registration) : undefined,
    fromRaw: (
      rawRegistration: RawRankingRegistration | undefined,
    ): RankingRegistration | undefined =>
      rawRegistration
        ? RawToRankingRegistrationMapper(rawRegistration)
        : undefined,
  },
}
