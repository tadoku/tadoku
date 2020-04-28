import { Medium, Language } from './interfaces'

export const AllMediums: Medium[] = [
  { id: 1, description: 'Book', unit: 'Pages' },
  { id: 2, description: 'Comic', unit: 'Pages' },
  { id: 3, description: 'Net', unit: 'Articles' },
  { id: 4, description: 'Full game', unit: 'Screens' },
  { id: 5, description: 'Game', unit: 'Screens' },
  { id: 6, description: 'Lyric', unit: 'Songs' },
  { id: 7, description: 'News', unit: 'Articles' },
  { id: 8, description: 'Sentences', unit: 'Sentences' },
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
export const mediumUnitById = (id: number) => MediumById[id].unit

export const AllLanguages: Language[] = [
  { code: 'ara', name: 'Arabic' },
  { code: 'zho', name: 'Chinese' },
  { code: 'hrv', name: 'Croatian' },
  { code: 'ces', name: 'Czech' },
  { code: 'nld', name: 'Dutch' },
  { code: 'eng', name: 'English' },
  { code: 'epo', name: 'Esperanto' },
  { code: 'fin', name: 'Finnish' },
  { code: 'fra', name: 'French' },
  { code: 'deu', name: 'German' },
  { code: 'ell', name: 'Greek' },
  { code: 'heb', name: 'Hebrew' },
  { code: 'gle', name: 'Irish' },
  { code: 'ita', name: 'Italian' },
  { code: 'jpn', name: 'Japanese' },
  { code: 'kor', name: 'Korean' },
  { code: 'pol', name: 'Polish' },
  { code: 'por', name: 'Portuguese' },
  { code: 'rus', name: 'Russian' },
  { code: 'spa', name: 'Spanish' },
  { code: 'swe', name: 'Swedish' },
  { code: 'tha', name: 'Thai' },
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

export const GlobalLanguage: Language = {
  code: 'GLO',
  name: 'Total',
}

export const languageNameByCode = (code: string) =>
  (LanguageByCode[code] || GlobalLanguage).name
