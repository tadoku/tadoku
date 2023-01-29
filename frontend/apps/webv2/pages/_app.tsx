import App, { AppProps } from 'next/app'
import { sdkServer as ory } from '@app/common/ory'
import { Atom, Provider } from 'jotai'
import { AppContextWithSession, sessionAtom } from '@app/common/session'
import { Session } from '@ory/client'
import { ToastContainer } from 'ui/components/toasts'
import 'ui/styles/globals.css'
import Navigation from '@app/ui/Navigation'
import { QueryCache, QueryClient, QueryClientProvider } from 'react-query'
import Head from 'next/head'
import Footer from '@app/ui/Footer'
import { Settings } from 'luxon'

// Default timezone for app
Settings.defaultZone = 'utc'

interface Props {
  session: Session | undefined
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

const MyApp = ({ Component, pageProps }: AppProps<Props>) => {
  const initialState = pageProps
  const { get: getInitialValues, set: setInitialValues } = createInitialValues()

  setInitialValues(sessionAtom, initialState.session)

  return (
    <Provider initialValues={getInitialValues()}>
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
        <div className="min-h-screen flex flex-col">
          <Navigation />
          <div className="p-4 md:px-8 md:pb-8 md:pt-4 mx-auto w-full max-w-7xl mb-auto">
            <Component {...pageProps} />
          </div>
          <Footer />
          <ToastContainer />
        </div>
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
