import { Medium, Language } from './interfaces'

export const AllMediums: Medium[] = [
  { id: 1, description: 'Book' },
  { id: 2, description: 'Manga' },
  { id: 3, description: 'Net' },
  { id: 4, description: 'Full game' },
  { id: 5, description: 'Game' },
  { id: 6, description: 'Lyric' },
  { id: 7, description: 'News' },
  { id: 8, description: 'Sentences' },
]

export const MediumById: { [key: number]: Medium } = AllMediums.reduce(
  (previous, current) => {
    return {
      ...previous,
      [current.id]: current,
    }
  },
  {},
)

export const mediumDescriptionById = (id: number) => MediumById[id].description

export const AllLanguages: Language[] = [
  { code: 'zho', name: 'Chinese' },
  { code: 'nld', name: 'Dutch' },
  { code: 'eng', name: 'English' },
  { code: 'fra', name: 'French' },
  { code: 'deu', name: 'German' },
  { code: 'ell', name: 'Greek' },
  { code: 'gle', name: 'Irish' },
  { code: 'ita', name: 'Italian' },
  { code: 'jpn', name: 'Japanese' },
  { code: 'kor', name: 'Korean' },
  { code: 'por', name: 'Portuguese' },
  { code: 'rus', name: 'Russian' },
  { code: 'spa', name: 'Spanish' },
  { code: 'swe', name: 'Swedish' },
  { code: 'tur', name: 'Turkish' },
]

export const LanguageByCode: { [key: string]: Language } = AllLanguages.reduce(
  (previous, current) => {
    return {
      ...previous,
      [current.code]: current,
    }
  },
  {},
)

export const languageNameByCode = (code: string) =>
  (LanguageByCode[code] || GlobalLanguage).name

export const GlobalLanguage: Language = {
  code: 'GLO',
  name: 'Total',
}
