import { RawMapper } from '../interfaces'

export interface Ranking {
  contestId: number
  userId: number
  userDisplayName: string
  languageCode: string
  amount: number
}

export interface rawRanking {
  contest_id: number
  user_id: number
  user_display_name: string
  language_code: string
  amount: number
}

export const RankingMapper: RawMapper<rawRanking, Ranking> = raw => ({
  contestId: raw.contest_id,
  userId: raw.user_id,
  userDisplayName: raw.user_display_name,
  languageCode: raw.language_code,
  amount: raw.amount,
})

export interface RankingRegistrationOverview {
  contestId: number
  userId: number
  userDisplayName: string

  registrations: {
    languageCode: string
    amount: number
  }[]
}

export interface RankingRegistration {
  start: Date
  end: Date
  contestId: number
  languages: string[]
}

export interface rawRankingRegistration {
  start: string
  end: string
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
  date: string
}

export interface AggregatedContestLogsByDayEntry {
  x: Date // day for x axis
  y: number // page count for y axis
}
