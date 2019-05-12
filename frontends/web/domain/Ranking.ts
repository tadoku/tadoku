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
