import { accountAppUrl, useAppUrls } from 'ui/app-urls'
import { useAccountApp } from './session'

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
  const urls = useAppUrls()
  const { homeUrl } = urls
  const accountApp = useAccountApp()
  const authUiUrl = accountAppUrl(urls, accountApp)
  return {
    ...routes,
    authSettings: (returnUrl = '') => `${authUiUrl}/${accountApp === 'v2' ? 'settings' : ''}?return_to=${encodeURIComponent(returnUrl)}`,
    authLogin: (returnUrl = '') => `${authUiUrl}/login?return_to=${encodeURIComponent(returnUrl)}`,
    mainApp: () => homeUrl,
  }
}
