/* eslint "@typescript-eslint/no-var-requires": "off" */
const withCss = require('@zeit/next-css')

require('dotenv').config()

module.exports = withCss({
  env: {
    GHOST_KEY: process.env.GHOST_KEY,
  },
})
