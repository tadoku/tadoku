require('dotenv').config()

const publicRuntimeConfig = {
  SESSION_COOKIE_NAME:
    process.env.NEXT_PUBLIC_SESSION_COOKIE_NAME || 'session_token',
  CHANGE_USERNAME_ENABLED:
    process.env.NEXT_PUBLIC_CHANGE_USERNAME_ENABLED || 'true',
}

module.exports = {
  publicRuntimeConfig,
}
