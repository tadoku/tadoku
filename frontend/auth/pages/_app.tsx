import App, { AppProps } from 'next/app'
import ory from '../src/ory'
import { Atom, Provider } from 'jotai'
import { AppContextWithSession, sessionAtom } from '../src/session'
import { Session } from '@ory/client'
import ToastContainer from 'tadoku-ui/components/toasts'
import 'tadoku-ui/styles/globals.css'
import Navbar from 'tadoku-ui/components/Navbar'
import {
  ArrowRightOnRectangleIcon,
  Cog8ToothIcon,
  UserIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/20/solid'

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
        <Navbar
          navigation={[
            { type: 'link', label: 'Home', href: '#', current: true },
            { type: 'link', label: 'Blog', href: '#', current: false },
            { type: 'link', label: 'Ranking', href: '#', current: false },
            { type: 'link', label: 'Manual', href: '#', current: false },
            { type: 'link', label: 'Forum', href: '#', current: false },
            {
              type: 'dropdown',
              label: 'John Doe',
              links: [
                {
                  label: 'Admin',
                  href: '#',
                  IconComponent: WrenchScrewdriverIcon,
                },
                { label: 'Settings', href: '#', IconComponent: Cog8ToothIcon },
                { label: 'Profile', href: '#', IconComponent: UserIcon },
                {
                  label: 'Log out',
                  href: '#',
                  IconComponent: ArrowRightOnRectangleIcon,
                },
              ],
            },
          ]}
        />
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
