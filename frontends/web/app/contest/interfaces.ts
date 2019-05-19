export interface Contest {
  id: number
  start: Date
  end: Date
  open: boolean
}

export interface rawContest {
  id: number
  start: string
  end: string
  open: boolean
}
