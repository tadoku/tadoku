import { get, post } from '../api'
import {
  Ranking,
  rawRanking,
  RankingRegistration,
  rawRankingRegistration,
  ContestLog,
  rawContestLog,
} from './interfaces'

const getRankings = async (contest_id?: number): Promise<Ranking[]> => {
  const response = await get(`/rankings?contest_id=${contest_id}`)

  if (response.status !== 200) {
    return []
  }

  const data: rawRanking[] = await response.json()

  return data.map(raw => ({
    userId: raw.user_id,
    userDisplayName: raw.user_display_name,
    languageCode: raw.language_code,
    amount: raw.amount,
  }))
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

  const data: rawRankingRegistration = await response.json()

  return {
    end: data.end,
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

  const data: rawRanking[] = await response.json()

  return data.map(raw => ({
    userId: raw.user_id,
    userDisplayName: raw.user_display_name,
    languageCode: raw.language_code,
    amount: raw.amount,
  }))
}

const createLog = async (payload: {
  contestId: number
  mediumId: number
  amount: number
  languageCode: string
}): Promise<boolean> => {
  const response = await post(`/contest_logs`, {
    body: {
      contest_id: payload.contestId,
      medium_id: payload.mediumId,
      amount: payload.amount,
      language_code: payload.languageCode,
    },
    authenticated: true,
  })

  return response.status === 201
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

  const data: rawContestLog[] = await response.json()

  return data.map(raw => ({
    id: raw.id,
    contestId: raw.contest_id,
    userId: raw.user_id,
    languageCode: raw.language_code,
    mediumId: raw.medium_id,
    amount: raw.amount,
    adjustedAmount: raw.adjusted_amount,
    date: new Date(raw.date),
  }))
}

const RankingApi = {
  get: getRankings,
  getCurrentRegistration,
  getRankingsRegistration,
  createLog,
  getLogsFor,
}

export default RankingApi
