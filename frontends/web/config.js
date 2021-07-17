const publicRuntimeConfig = {
  GHOST_KEY: process.env.GHOST_KEY || '22c9b1088957389622d210662f',
  GHOST_URL: process.env.GHOST_URL || 'https://blog.tadoku.app',
  SESSION_COOKIE_NAME: process.env.SESSION_COOKIE_NAME,
}

module.exports = publicRuntimeConfig
