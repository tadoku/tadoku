/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  publicRuntimeConfig: {
    kratosPublicEndpoint: process.env.NEXT_PUBLIC_KRATOS_ENDPOINT,
    kratosInternalEndpoint: process.env.NEXT_PUBLIC_KRATOS_INTERNAL_ENDPOINT,
  }
}

const withTM = require('next-transpile-modules')(['ui']);

module.exports = withTM(nextConfig)
