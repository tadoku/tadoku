import { Ranking, RankingWithRank } from '../interfaces'

export const aggregateRankingLeaderboard = (
  rankings: Ranking[],
): RankingWithRank[] => {
  const initialRankingsByScore: { [key: number]: Ranking[] } = {}

  const rankingsByScore = rankings.reduce((scores, ranking) => {
    if (!scores[ranking.amount]) {
      scores[ranking.amount] = []
    }

    scores[ranking.amount].push(ranking)

    return scores
  }, initialRankingsByScore)

  const sortedScores = Object.keys(rankingsByScore)
    .map(n => Number(n))
    .sort((a, b) => b - a)

  const initialResult: {
    currentRank: number
    rankings: RankingWithRank[]
  } = {
    currentRank: 1,
    rankings: [],
  }

  const result = sortedScores.reduce((result, score) => {
    const rankings = rankingsByScore[score]

    const newRankings = rankings.map(ranking => ({
      data: ranking,
      tied: rankings.length > 1,
      rank: result.currentRank,
    }))

    result.rankings.push(...newRankings)
    result.currentRank += rankings.length

    return result
  }, initialResult)

  return result.rankings
}
