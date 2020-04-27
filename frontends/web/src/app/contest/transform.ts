import { Contest, RawContest } from './interfaces'
import { Mapper, Mappers } from '../interfaces'
import { Serializer } from '../cache'
import { createSerializer, createMappers } from '../transform'

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

export const contestSerializer = createSerializer(contestMapper)

export const contestsSerializer: Serializer<Contest[]> = {
  serialize: contests => {
    const raw = contests.map(contestToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(rawToContestMapper)
  },
}
