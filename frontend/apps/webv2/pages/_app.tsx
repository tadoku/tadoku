import App, { AppProps } from 'next/app'
import { NextPage } from 'next'
import { ReactElement, ReactNode } from 'react'
import { Atom, Provider } from 'jotai'
import {
  AppContextWithSession,
  sessionAtom,
  accountAppAtom,
  useUserRole,
} from '@app/common/session'
import { Account, AccountApp, readIdentitySession } from 'ui/identity'
import { ToastContainer } from 'ui/components/toasts'
import 'ui/styles/globals.css'
import Navigation from '@app/ui/Navigation'
import AnnouncementBanner from '@app/ui/AnnouncementBanner'
import BannedScreen from '@app/ui/BannedScreen'
import { QueryCache, QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Head from 'next/head'
import Footer from '@app/ui/Footer'
import { Settings } from 'luxon'
import {
  defaultFeatureFlagDecisions,
  FeatureFlagDecisions,
} from '@app/feature-flags/registry'
import {
  FeatureFlagRefresh,
  featureFlagDecisionsAtom,
} from '@app/feature-flags/client'
import { bootstrapFeatureFlagDecisions } from '@app/feature-flags/bootstrap'
import { AppUrls, AppUrlsProvider, appUrlsForHost, serverApiEndpointForHost } from 'ui/app-urls'

// Default timezone for app
Settings.defaultZone = 'utc'

interface Props {
  appUrls: AppUrls
  session: Account | undefined
  accountApp: AccountApp
  featureFlags: FeatureFlagDecisions
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

const AppContent = ({ children }: { children: ReactNode }) => {
  const role = useUserRole()

  if (role === 'banned') {
    return <BannedScreen />
  }

  return <>{children}</>
}

const MyApp = ({ Component, pageProps }: AppPropsWithLayout) => {
  const initialState = pageProps
  const { get: getInitialValues, set: setInitialValues } = createInitialValues()

  setInitialValues(sessionAtom, initialState.session)
  setInitialValues(accountAppAtom, initialState.accountApp)
  setInitialValues(
    featureFlagDecisionsAtom,
    initialState.featureFlags ?? defaultFeatureFlagDecisions,
  )

  const getLayout = Component.getLayout ?? (page => page)

  return (
    <AppUrlsProvider urls={pageProps.appUrls}>
      <Provider key={`${initialState.accountApp}:${initialState.session?.id ?? 'guest'}`} initialValues={getInitialValues()}>
        <QueryClientProvider client={queryClient}>
          <Head>
            <title>Tadoku</title>
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
          <FeatureFlagRefresh>
            <AppContent>
              <div className="min-h-screen flex flex-col">
                <Navigation />
                <div className="p-4 md:px-8 md:pb-8 md:pt-4 mx-auto w-full max-w-7xl mb-auto">
                  <AnnouncementBanner />
                  {getLayout(<Component {...pageProps} />)}
                </div>
                <Footer />
                <ToastContainer />
              </div>
            </AppContent>
          </FeatureFlagRefresh>
        </QueryClientProvider>
      </Provider>
    </AppUrlsProvider>
  )
}

MyApp.getInitialProps = async (ctx: AppContextWithSession) => {
  ctx.ctx.res?.setHeader('Cache-Control', 'private, no-store')
  const cookie = ctx.ctx.req?.headers.cookie
  const host = ctx.ctx.req?.headers.host ?? (typeof window === 'undefined' ? undefined : window.location.host)
  const identity = await readIdentitySession(host, ctx.ctx.req?.headers)
  ctx.ctx.session = identity.account ?? undefined

  const initialAppProps = await App.getInitialProps(ctx)
  initialAppProps.pageProps.session = ctx.ctx.session
  initialAppProps.pageProps.accountApp = identity.account_app
  initialAppProps.pageProps.appUrls = appUrlsForHost(host)
  initialAppProps.pageProps.featureFlags = { ...defaultFeatureFlagDecisions }

  initialAppProps.pageProps.featureFlags = await bootstrapFeatureFlagDecisions(
    ctx.ctx.session,
    cookie,
    typeof window === 'undefined',
    fetch,
    3_000,
    serverApiEndpointForHost(host),
    ctx.ctx.req?.headers['x-dev-branch'],
  )

  return initialAppProps
}

export default MyApp
