import { get, post, put, destroy } from '../api'
import {
  Ranking,
  RawRanking,
  RankingRegistration,
  RawRankingRegistration,
  ContestLog,
  RawContestLog,
} from './interfaces'
import { rankingMapper } from './transform/ranking'
import { contestLogMapper } from './transform/contest-log'

const joinContest = async (
  contestId: number,
  languages: string[],
): Promise<boolean> => {
  const response = await post(`/rankings`, {
    body: {
      contest_id: contestId,
      languages: languages,
    },
    authenticated: true,
  })

  return response.status === 201
}

const getRankings = async (contestId?: number): Promise<Ranking[]> => {
  const response = await get(`/rankings?contest_id=${contestId}`)

  if (response.status !== 200) {
    return []
  }

  const data: RawRanking[] = await response.json()

  return data.map(rankingMapper.fromRaw)
}

const getCurrentRegistration = async (): Promise<
  RankingRegistration | undefined
> => {
  const response = await get('/rankings/current', {
    authenticated: true,
  })

  if (response.status != 200) {
    return undefined
  }

  const data: RawRankingRegistration = await response.json()

  return {
    start: new Date(data.start),
    end: new Date(data.end),
    contestId: data.contest_id,
    languages: data.languages,
  }
}

const getRankingsRegistration = async (
  contestId: number,
  userId: number,
): Promise<Ranking[]> => {
  const response = await get(
    `/rankings/registration?contest_id=${contestId}&user_id=${userId}`,
  )

  if (response.status !== 200) {
    return []
  }

  const data: RawRanking[] = await response.json()

  return data.map(rankingMapper.fromRaw)
}

const createLog = async (payload: {
  contestId: number
  mediumId: number
  amount: number
  languageCode: string
  description: string
}): Promise<boolean> => {
  const response = await post(`/contest_logs`, {
    body: {
      contest_id: payload.contestId,
      medium_id: payload.mediumId,
      amount: payload.amount,
      language_code: payload.languageCode,
      description: payload.description,
    },
    authenticated: true,
  })

  return response.status === 201
}

const deleteLog = async (contestId: number): Promise<boolean> => {
  const response = await destroy(`/contest_logs/${contestId}`, {
    authenticated: true,
  })

  return response.status === 200
}

const updateLog = async (
  id: number,
  payload: {
    contestId: number
    mediumId: number
    amount: number
    languageCode: string
    description: string
  },
): Promise<boolean> => {
  const response = await put(`/contest_logs/${id}`, {
    body: {
      contest_id: payload.contestId,
      medium_id: payload.mediumId,
      amount: payload.amount,
      language_code: payload.languageCode,
      description: payload.description,
    },
    authenticated: true,
  })

  return response.status === 204
}

const getLogsFor = async (
  contestId: number,
  userId: number,
): Promise<ContestLog[]> => {
  const response = await get(
    `/contest_logs?contest_id=${contestId}&user_id=${userId}`,
  )

  if (response.status != 200) {
    return []
  }

  const data: RawContestLog[] = await response.json()

  return data.map(contestLogMapper.fromRaw)
}

const RankingApi = {
  joinContest,
  get: getRankings,
  getCurrentRegistration,
  getRankingsRegistration,
  createLog,
  deleteLog,
  updateLog,
  getLogsFor,
}

export default RankingApi
