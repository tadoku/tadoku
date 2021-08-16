jest.mock('next/config', () => () => ({
  publicRuntimeConfig: {
    CHANGE_USERNAME_ENABLED: 'true',
  },
}))
