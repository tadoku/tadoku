import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()
const kratos = publicRuntimeConfig.authUiUrl
const homeUrl = publicRuntimeConfig.homeUrl

export const routes = {
  home: () => `/`,
  posts: () => `/posts`,
  postNew: () => `/posts/new`,
  postPreview: (slug: string) => `/posts/${slug}`,
  postEdit: (slug: string) => `/posts/${slug}/edit`,
  pages: () => `/pages`,
  users: () => `/users`,

  // External
  authSettings: (return_url?: string) =>
    `${kratos}/?return_to=${return_url ?? ''}`,
  authLogin: (return_url?: string) =>
    `${kratos}/login?return_to=${return_url ?? ''}`,
  mainApp: () => homeUrl,
}
