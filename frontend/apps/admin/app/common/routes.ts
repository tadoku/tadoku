import { useAppUrls } from 'ui/app-urls'

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

}

export function useRoutes() {
  const { authUiUrl, homeUrl } = useAppUrls()
  return {
    ...routes,
    authSettings: (returnUrl = '') => `${authUiUrl}/?return_to=${returnUrl}`,
    authLogin: (returnUrl = '') => `${authUiUrl}/login?return_to=${returnUrl}`,
    mainApp: () => homeUrl,
  }
}
