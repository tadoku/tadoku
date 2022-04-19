export interface Contest {
  id: number
  description: string
  start: Date
  end: Date
  open: boolean
}

export interface RawContest {
  id: number
  description: string
  start: string
  end: string
  open: boolean
}

export interface ContestStats {
  byLanguage: {
    languageCode: string
    count: number
  }[]
  participants: number
  totalPages: number
}
