import { Ranking, RawRanking } from './../interfaces'
import { Mapper, Mappers } from '../../interfaces'
import { Serializer } from '../../cache'
import { createMappers } from '../../transform'

const rawToRankingMapper: Mapper<RawRanking, Ranking> = raw => ({
  contestId: raw.contest_id,
  userId: raw.user_id,
  userDisplayName: raw.user_display_name,
  languageCode: raw.language_code,
  amount: raw.amount,
})

const rankingToRawMapper: Mapper<Ranking, RawRanking> = ranking => ({
  contest_id: ranking.contestId,
  user_id: ranking.userId,
  user_display_name: ranking.userDisplayName,
  language_code: ranking.languageCode,
  amount: ranking.amount,
})

export const rankingsSerializer: Serializer<Ranking[]> = {
  serialize: data => {
    const raw = data.map(rankingToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(rawToRankingMapper)
  },
}

export const rankingMapper: Mappers<RawRanking, Ranking> = createMappers({
  toRaw: rankingToRawMapper,
  fromRaw: rawToRankingMapper,
})
