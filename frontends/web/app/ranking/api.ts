import { get } from '../../domain/api/api'
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

const RankingApi = {
  get: getRankings,
  getCurrentRegistration,
}

export default RankingApi
