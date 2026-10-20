import App, { AppProps } from 'next/app'
import { sdkServer as ory } from '../src/ory'
import { Atom, Provider } from 'jotai'
import { AppContextWithSession, sessionAtom } from '../src/session'
import { Session } from '@ory/kratos-client'
import { ToastContainer } from 'ui'
import 'ui/styles/globals.css'
import Navigation from '../src/Navigation'
import Head from 'next/head'
import { readIdentitySession } from 'ui/identity'
import { SessionUnavailable } from 'ui/components/SessionUnavailable'
import { AppUrls, AppUrlsProvider, appUrlsForHost, accountAppUrl } from 'ui/app-urls'


interface Props {
  appUrls: AppUrls
  session: Session | undefined
  sessionUnavailable?: boolean
  redirecting?: boolean
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

  if (pageProps.redirecting) return null
  if (pageProps.sessionUnavailable) return <SessionUnavailable />

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
  ctx.ctx.res?.setHeader('Cache-Control', 'private, no-store')
  const host = ctx.ctx.req?.headers.host ?? (typeof window === 'undefined' ? undefined : window.location.host)
  const appUrls = appUrlsForHost(host)
  const identity = await readIdentitySession(host, ctx.ctx.req?.headers)
  if (!identity) return { pageProps: { appUrls, sessionUnavailable: true } }
  if (identity.account_app === 'v2') {
    const current = new URL(ctx.ctx.asPath ?? '/', appUrls.authUiUrl)
    const target = new URL(accountAppUrl(appUrls, 'v2'))
    target.pathname = current.pathname === '/' ? '/settings' : current.pathname === '/account-recovery' ? '/recovery' : current.pathname
    target.search = current.search
    target.hash = current.hash
    if (ctx.ctx.res) {
      ctx.ctx.res.writeHead(307, { Location: target.toString() })
      ctx.ctx.res.end()
    } else {
      window.location.replace(target.toString())
    }
    return { pageProps: { appUrls, redirecting: true } }
  }
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
    } catch {}
  }

  const initialAppProps = await App.getInitialProps(ctx)
  initialAppProps.pageProps.session = ctx.ctx.session
  initialAppProps.pageProps.appUrls = appUrls

  return { ...props, ...initialAppProps }
}

export default MyApp
