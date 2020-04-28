import { globalLanguage, mediaByID, languageByCode } from '../database'

export const formatMediaDescription = (id: number) => mediaByID[id].description
export const formatMediaUnit = (id: number) => mediaByID[id].unit

export const formatLanguageName = (code: string) =>
  (languageByCode[code] || globalLanguage).name

export const formatScore = (amount: number) => Math.round(amount * 10) / 10

export const scoreLabel = (languageCode: string) => {
  if (languageCode == globalLanguage.code) {
    return 'Overall score'
  }

  return `Score for ${formatLanguageName(languageCode)}`
}

export const formatPoints = (amount: number): string => {
  switch (amount) {
    case 0:
      return 'No points'
    case 1:
      return '1 point'
    default:
      return `${formatScore(amount)} points`
  }
}
