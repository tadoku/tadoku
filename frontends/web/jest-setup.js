jest.mock('next/config', () => () => ({
  publicRuntimeConfig: require('./config'),
}))
