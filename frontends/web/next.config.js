/* eslint "@typescript-eslint/no-var-requires": "off" */
const withCss = require('@zeit/next-css')

require('dotenv').config()

module.exports = withCss({
  env: {
    API_ROOT: process.env.API_ROOT,
    GHOST_KEY: process.env.GHOST_KEY,
    GHOST_URL: process.env.GHOST_URL,
  },
})
