import { get } from './api'
import { Ranking, rawRanking } from '../Ranking'

const getRankings = async (contest_id?: number): Promise<Ranking[]> => {
  const response = await get(`/rankings?contest_id=${contest_id}`)

  if (response.status !== 200) {
    return []
  }

  const data: rawRanking[] = await response.json()

  return data.map(raw => ({
    userId: raw.user_id,
    languageCode: raw.language_code,
    amount: raw.amount,
  }))
}

const RankingApi = {
  get: getRankings,
}

export default RankingApi
