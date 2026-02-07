import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()
const kratos = publicRuntimeConfig.authUiUrl
const homeUrl = publicRuntimeConfig.homeUrl

export const routes = {
  home: () => `/`,
  posts: (ns: string) => `/posts/${ns}`,
  postNew: (ns: string) => `/posts/${ns}/new`,
  postPreview: (ns: string, id: string) => `/posts/${ns}/${id}`,
  postEdit: (ns: string, id: string) => `/posts/${ns}/${id}/edit`,
  pages: (ns: string) => `/pages/${ns}`,
  pageNew: (ns: string) => `/pages/${ns}/new`,
  pagePreview: (ns: string, id: string) => `/pages/${ns}/${id}`,
  pageEdit: (ns: string, id: string) => `/pages/${ns}/${id}/edit`,
  announcements: (ns: string) => `/announcements/${ns}`,
  announcementNew: (ns: string) => `/announcements/${ns}/new`,
  announcementEdit: (ns: string, id: string) => `/announcements/${ns}/${id}/edit`,
  users: () => `/users`,
  languages: () => `/languages`,

  // External
  authSettings: (return_url?: string) =>
    `${kratos}/?return_to=${return_url ?? ''}`,
  authLogin: (return_url?: string) =>
    `${kratos}/login?return_to=${return_url ?? ''}`,
  mainApp: () => homeUrl,
}
