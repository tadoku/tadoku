import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()
const kratos = publicRuntimeConfig.authUiUrl
const homeUrl = publicRuntimeConfig.homeUrl

export const routes = {
  home: () => `/`,
  posts: () => `/posts`,
  postNew: () => `/posts/new`,
  postPreview: (id: string) => `/posts/${id}`,
  postEdit: (id: string) => `/posts/${id}/edit`,
  pages: () => `/pages`,
  pageNew: () => `/pages/new`,
  pagePreview: (id: string) => `/pages/${id}`,
  pageEdit: (id: string) => `/pages/${id}/edit`,
  users: () => `/users`,

  // External
  authSettings: (return_url?: string) =>
    `${kratos}/?return_to=${return_url ?? ''}`,
  authLogin: (return_url?: string) =>
    `${kratos}/login?return_to=${return_url ?? ''}`,
  mainApp: () => homeUrl,
}
