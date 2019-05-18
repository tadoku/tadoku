import { get, post } from '../api'
import {
  Ranking,
  rawRanking,
  RankingRegistration,
  rawRankingRegistration,
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

const RankingApi = {
  get: getRankings,
  getCurrentRegistration,
  createLog,
}

export default RankingApi
