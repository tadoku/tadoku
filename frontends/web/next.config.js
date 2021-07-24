require('dotenv').config()

const publicRuntimeConfig = {
  // TODO: Remove ghost environment
  GHOST_KEY: process.env.GHOST_KEY || '22c9b1088957389622d210662f',
  GHOST_URL: process.env.GHOST_URL || 'https://blog.tadoku.app',
  SESSION_COOKIE_NAME:
    process.env.NEXT_PUBLIC_SESSION_COOKIE_NAME || 'session_token',
  CHANGE_USERNAME_ENABLED:
    process.env.NEXT_PUBLIC_CHANGE_USERNAME_ENABLED || 'true',
}

module.exports = {
  publicRuntimeConfig,
}
