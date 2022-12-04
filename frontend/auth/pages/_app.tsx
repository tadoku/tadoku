import App, { AppProps } from 'next/app'
import ory from '../src/ory'
import { Atom, Provider } from 'jotai'
import { AppContextWithSession, sessionAtom } from '../src/session'
import { Session } from '@ory/client'
import Header from '../ui/Header'
import { ToastContainer } from 'react-toastify'
import 'react-toastify/dist/ReactToastify.css'

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
        <Header />
        <hr />
        <Component {...pageProps} />
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
    } catch (err) {}
  }

  const initialAppProps = await App.getInitialProps(ctx)
  initialAppProps.pageProps.session = ctx.ctx.session

  return { ...props, ...initialAppProps }
}

export default MyApp
