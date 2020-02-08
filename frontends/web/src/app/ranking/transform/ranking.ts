import { Ranking, RawRanking } from './../interfaces'
import { Mapper } from '../../interfaces'
import { Serializer } from '../../cache'

const RawToRankingMapper: Mapper<RawRanking, Ranking> = raw => ({
  contestId: raw.contest_id,
  userId: raw.user_id,
  userDisplayName: raw.user_display_name,
  languageCode: raw.language_code,
  amount: raw.amount,
})

const RankingToRawMapper: Mapper<Ranking, RawRanking> = ranking => ({
  contest_id: ranking.contestId,
  user_id: ranking.userId,
  user_display_name: ranking.userDisplayName,
  language_code: ranking.languageCode,
  amount: ranking.amount,
})

export const RankingsSerializer: Serializer<Ranking[]> = {
  serialize: data => {
    const raw = data.map(RankingToRawMapper)
    return JSON.stringify(raw)
  },
  deserialize: serializedData => {
    let raw = JSON.parse(serializedData)
    return raw.map(RawToRankingMapper)
  },
}

export const RankingMapper = {
  toRaw: RankingToRawMapper,
  fromRaw: RawToRankingMapper,
}
