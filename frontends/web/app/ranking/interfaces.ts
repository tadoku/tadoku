export interface Ranking {
  userId: number
  userDisplayName: string
  languageCode: string
  amount: number
}

export interface rawRanking {
  user_id: number
  user_display_name: string
  language_code: string
  amount: number
}

export interface RankingRegistration {
  end: Date
  contestId: number
  languages: string[]
}

export interface rawRankingRegistration {
  end: Date
  contest_id: number
  languages: string[]
}

export interface Medium {
  id: number
  description: string
}

export interface Language {
  code: string
  name: string
}

export interface ContestLog {
  id: number
  contestId: number
  userId: number
  languageCode: string
  mediumId: number
  amount: number
  adjustedAmount: number
  date: Date
}

export interface rawContestLog {
  id: number
  contest_id: number
  user_id: number
  language_code: string
  medium_id: number
  amount: number
  adjusted_amount: number
  date: Date
}
