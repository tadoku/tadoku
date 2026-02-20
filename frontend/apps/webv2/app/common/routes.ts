import { DateTime } from 'luxon'
import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()
const kratos = publicRuntimeConfig.authUiUrl
const adminUrl = publicRuntimeConfig.adminUrl

type Page = string | number

export const routes = {
  home: () => `/`,

  blogPost: (slug: string) => `/blog/posts/${slug}`,
  blogList: (page?: Page) => `/blog/${page ?? '1'}`,
  blogPage: (slug: string) => `/pages/${slug}`,

  contestListOfficial: (page?: Page) => `/contests/official/${page ?? '1'}`,
  contestListUserContests: (page?: Page) =>
    `/contests/user-contests/${page ?? '1'}`,
  contestListMyContests: (page?: Page) =>
    `/contests/my-contests/${page ?? '1'}`,
  contestNew: () => `/contests/new`,
  contestLeaderboard: (id: string, page?: Page) =>
    `/contests/${id}/leaderboard/${page ?? '1'}`,
  contestUpdates: (id: string, page?: Page) =>
    `/contests/${id}/updates/${page ?? '1'}`,
  contestJoin: (id: string) => `/contests/${id}/registration`,
  contestUserProfile: (contest_id: string, user_id: string, page?: Page) =>
    `/contests/${contest_id}/profile/${user_id}/${page ?? '1'}`,

  logCreate: () => `/logs/new`,
  logCreateWithAmount: (amount: number) => `/logs/new?amount=${amount}`,
  log: (id: string) => `/logs/${id}`,
  logEdit: (id: string) => `/logs/${id}/edit`,
  logContests: (id: string) => `/logs/${id}/contests`,

  leaderboardLatestOfficial: (page?: Page) =>
    `/leaderboard/latest/${page ?? '1'}`,
  leaderboardAllTimeOfficial: (page?: Page) =>
    `/leaderboard/all-time/${page ?? '1'}`,
  leaderboardYearlyOfficial: (year?: Page, page?: Page) =>
    `/leaderboard/yearly/${year ?? DateTime.now().year}/${page ?? '1'}`,

  userProfileStatistics: (id: string, year?: Page) =>
    `/profile/${id}/statistics/${year ?? DateTime.now().year}`,
  userProfileUpdates: (id: string, page?: Page) =>
    `/profile/${id}/updates/${page ?? '1'}`,

  authSettings: (return_url?: string) =>
    `${kratos}/?return_to=${return_url ?? ''}`,
  authLogin: (return_url?: string) =>
    `${kratos}/login?return_to=${return_url ?? ''}`,
  authSignup: (return_url?: string) =>
    `${kratos}/register?return_to=${return_url ?? ''}`,

  pageCounter: () => `/page-counter`,

  // External

  admin: () => adminUrl,

  personalWebsite: () => `https://antonve.be`,
  twitter: () => `https://twitter.com/tadoku_app`,
  github: () => `https://github.com/tadoku`,
  discord: () => `https://discord.gg/Dd8t9WB`,
}
