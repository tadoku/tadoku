import Router from 'next/router'

export const refresh = () => {
  if (Router.asPath) {
    Router.push(Router.asPath)
  }
}
