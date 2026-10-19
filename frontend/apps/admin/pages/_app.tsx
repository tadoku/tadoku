import App, { AppProps } from 'next/app'
import { NextPage } from 'next'
import { ReactElement, ReactNode, useEffect } from 'react'
import { Atom, Provider, useAtom } from 'jotai'
import {
  AppContextWithSession,
  sessionAtom,
  accountAppAtom,
  useUserRole,
  useCurrentLocation,
} from '@app/common/session'
import { Account, AccountApp, readIdentitySession } from 'ui/identity'
import { ToastContainer } from 'ui/components/toasts'
import 'ui/styles/globals.css'
import { QueryCache, QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Head from 'next/head'
import { Settings } from 'luxon'
import AccessDenied from '@app/ui/AccessDenied'
import LoadingScreen from '@app/ui/LoadingScreen'
import { AppUrls, AppUrlsProvider, appUrlsForHost } from 'ui/app-urls'
import { useRoutes } from '@app/common/routes'

// Default timezone for app
Settings.defaultZone = 'utc'

interface Props {
  appUrls: AppUrls
  session: Account | undefined
  accountApp: AccountApp
}

export type NextPageWithLayout<P = {}, IP = P> = NextPage<P, IP> & {
  getLayout?: (page: ReactElement) => ReactNode
}

type AppPropsWithLayout = AppProps<Props> & {
  Component: NextPageWithLayout<Props>
}

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: error => {
      if ((error as Error).message === '401') {
        window.location.reload()
      }
    },
  }),
})

const createInitialValues = () => {
  const initialValues: (readonly [Atom<unknown>, unknown])[] = []
  const get = () => initialValues
  const set = function <Value>(anAtom: Atom<Value>, value: Value) {
    initialValues.push([anAtom, value])
  }
  return { get, set }
}

const RedirectToLogin = () => {
  const routes = useRoutes()
  const loginUrl = routes.authLogin(useCurrentLocation())

  useEffect(() => {
    window.location.href = loginUrl
  }, [loginUrl])

  return <LoadingScreen />
}

const AppContent = ({ children }: { children: ReactNode }) => {
  const [session] = useAtom(sessionAtom)
  const role = useUserRole()

  // No session - redirect to login
  if (!session) {
    return <RedirectToLogin />
  }

  // Still loading role
  if (role === undefined) {
    return <LoadingScreen />
  }

  // Not an admin
  if (role !== 'admin') {
    return <AccessDenied />
  }

  return <>{children}</>
}

const MyApp = ({ Component, pageProps }: AppPropsWithLayout) => {
  const initialState = pageProps
  const { get: getInitialValues, set: setInitialValues } = createInitialValues()

  setInitialValues(sessionAtom, initialState.session)
  setInitialValues(accountAppAtom, initialState.accountApp)

  const getLayout = Component.getLayout ?? (page => page)

  return (
    <AppUrlsProvider urls={pageProps.appUrls}>
      <Provider key={`${initialState.accountApp}:${initialState.session?.id ?? 'guest'}`} initialValues={getInitialValues()}>
        <QueryClientProvider client={queryClient}>
          <Head>
            <title>Admin - Tadoku</title>
            <link
              href="/favicon.png"
              rel="shortcut icon"
              media="(prefers-color-scheme: light)"
            />
            <link
              href="/favicon-dark.png"
              rel="shortcut icon"
              media="(prefers-color-scheme: dark)"
            />
          </Head>
          <AppContent>
            {getLayout(<Component {...pageProps} />)}
            <ToastContainer />
          </AppContent>
        </QueryClientProvider>
      </Provider>
    </AppUrlsProvider>
  )
}

MyApp.getInitialProps = async (ctx: AppContextWithSession) => {
  ctx.ctx.res?.setHeader('Cache-Control', 'private, no-store')
  const host = ctx.ctx.req?.headers.host ?? (typeof window === 'undefined' ? undefined : window.location.host)
  const identity = await readIdentitySession(host, ctx.ctx.req?.headers)
  ctx.ctx.session = identity.account ?? undefined

  const initialAppProps = await App.getInitialProps(ctx)
  initialAppProps.pageProps.session = ctx.ctx.session
  initialAppProps.pageProps.accountApp = identity.account_app
  initialAppProps.pageProps.appUrls = appUrlsForHost(ctx.ctx.req?.headers.host ?? (typeof window === 'undefined' ? undefined : window.location.host))

  return initialAppProps
}

export default MyApp
