/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  publicRuntimeConfig: {
    kratosPublicEndpoint: process.env.NEXT_PUBLIC_KRATOS_ENDPOINT,
    kratosInternalEndpoint: process.env.NEXT_PUBLIC_KRATOS_INTERNAL_ENDPOINT,
    authUiUrl: process.env.NEXT_PUBLIC_AUTH_UI_URL,
    homeUrl: process.env.NEXT_PUBLIC_HOME_URL,
  }
}

const withTM = require('next-transpile-modules')(['tadoku-ui']);

module.exports = withTM(nextConfig)
