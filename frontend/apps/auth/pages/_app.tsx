import App, { AppProps } from 'next/app'
import { sdkServer as ory } from '../src/ory'
import { Atom, Provider } from 'jotai'
import { AppContextWithSession, sessionAtom } from '../src/session'
import { Session } from '@ory/client'
import ToastContainer from 'tadoku-ui/components/toasts'
import 'tadoku-ui/styles/globals.css'
import Navigation from '../src/Navigation'

interface Props {
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
    <Provider initialValues={getInitialValues()}>
      <div>
        <Navigation />
        <div className="p-8 mx-auto max-w-xl">
          <Component {...pageProps} />
        </div>
        <ToastContainer />
      </div>
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
    } catch (err) {
      ctx.ctx.res?.setHeader('Set-Cookie', ['ory_kratos_session=0; Max-Age=0'])
    }
  }

  const initialAppProps = await App.getInitialProps(ctx)
  initialAppProps.pageProps.session = ctx.ctx.session

  return { ...props, ...initialAppProps }
}

export default MyApp
