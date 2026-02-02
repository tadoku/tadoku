import App, { AppProps } from 'next/app'
import { NextPage } from 'next'
import { ReactElement, ReactNode, useEffect } from 'react'
import { sdkServer as ory } from '@app/common/ory'
import { Atom, Provider } from 'jotai'
import { AppContextWithSession, sessionAtom, useUserRole, useCurrentLocation } from '@app/common/session'
import { Session } from '@ory/client'
import { ToastContainer } from 'ui/components/toasts'
import 'ui/styles/globals.css'
import { QueryCache, QueryClient, QueryClientProvider } from 'react-query'
import Head from 'next/head'
import { Settings } from 'luxon'
import AccessDenied from '@app/ui/AccessDenied'
import LoadingScreen from '@app/ui/LoadingScreen'
import { routes } from '@app/common/routes'

// Default timezone for app
Settings.defaultZone = 'utc'

interface Props {
  session: Session | undefined
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
        window.location.pathname = '/api/unauthorized'
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
  const currentUrl = useCurrentLocation()

  useEffect(() => {
    window.location.href = routes.authLogin(currentUrl)
  }, [currentUrl])

  return <LoadingScreen />
}

const AppContent = ({ children }: { children: ReactNode }) => {
  const role = useUserRole()

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

  const getLayout = Component.getLayout ?? ((page) => page)

  // If no session, redirect to login
  if (!initialState.session) {
    return (
      <Provider initialValues={getInitialValues()}>
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
        <RedirectToLogin />
      </Provider>
    )
  }

  return (
    <Provider initialValues={getInitialValues()}>
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
  )
}

MyApp.getInitialProps = async (ctx: AppContextWithSession) => {
  const cookie = ctx.ctx.req?.headers.cookie
  const props = {
    pageProps: {
      initialState: {
        session: undefined as Session | undefined,
      },
    },
  }

  if (cookie) {
    try {
      const { data: session } = await ory.toSession(undefined, cookie)
      props.pageProps.initialState.session = session
      ctx.ctx.session = session
    } catch (err) {}
  }

  const initialAppProps = await App.getInitialProps(ctx)
  initialAppProps.pageProps.session = ctx.ctx.session

  return { ...props, ...initialAppProps }
}

export default MyApp
