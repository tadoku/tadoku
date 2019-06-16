import { Contest, rawContest } from './interfaces'
import { Mapper } from '../interfaces'
import { Serializer } from '../cache'

export const RawToContestMapper: Mapper<rawContest, Contest> = raw => ({
  id: raw.id,
  description: raw.description,
  start: new Date(raw.start),
  end: new Date(raw.end),
  open: raw.open,
})

export const ContestToRawMapper: Mapper<Contest, rawContest> = contest => ({
  id: contest.id,
  description: contest.description,
  start: contest.start.toISOString(),
  end: contest.end.toISOString(),
  open: contest.open,
})

export const ContestSerializer: Serializer<Contest> = {
  serialize: contest => {
    const raw = ContestToRawMapper(contest)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return RawToContestMapper(raw)
  },
}
