import App, { AppProps } from 'next/app'
import { sdkServer as ory } from '../src/ory'
import { Atom, Provider } from 'jotai'
import { AppContextWithSession, sessionAtom } from '../src/session'
import { Session } from '@ory/kratos-client'
import { ToastContainer } from 'ui'
import 'ui/styles/globals.css'
import Navigation from '../src/Navigation'
import Head from 'next/head'
import getConfig from 'next/config'
import { AppUrls, AppUrlsProvider, appUrlsForHost } from 'ui/app-urls'

const { publicRuntimeConfig } = getConfig()

interface Props {
  appUrls: AppUrls
  session: Session | undefined
}

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
    <AppUrlsProvider urls={pageProps.appUrls}>
      <Provider initialValues={getInitialValues()}>
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
        <div>
          <Navigation />
          <div className="p-8 mx-auto max-w-xl">
            <Component {...pageProps} />
          </div>
          <ToastContainer />
        </div>
      </Provider>
    </AppUrlsProvider>
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
      const { data: session } = await ory.toSession({ cookie })
      props.pageProps.initialState.session = session
      ctx.ctx.session = session
    } catch (err) {
      const cookieAttributes = [
        'ory_kratos_session=0',
        'Path=/',
        `Domain=${publicRuntimeConfig.cookieDomain}`,
        'Max-Age=0',
        'SameSite=Lax',
      ]
      if (publicRuntimeConfig.cookieSecure) {
        cookieAttributes.push('Secure')
      }
      ctx.ctx.res?.setHeader('Set-Cookie', [cookieAttributes.join('; ')])
    }
  }

  const initialAppProps = await App.getInitialProps(ctx)
  initialAppProps.pageProps.session = ctx.ctx.session
  initialAppProps.pageProps.appUrls = appUrlsForHost(ctx.ctx.req?.headers.host ?? (typeof window === 'undefined' ? undefined : window.location.host))

  return { ...props, ...initialAppProps }
}

export default MyApp
