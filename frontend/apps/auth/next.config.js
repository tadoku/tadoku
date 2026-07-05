/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  publicRuntimeConfig: {
    // TODO: Figure out why this isn't getting passed to the client despite being opted-out from automatic static optimization
    kratosPublicEndpoint:
      process.env.NEXT_PUBLIC_KRATOS_ENDPOINT ??
      'https://account.tadoku.app/kratos',
    kratosInternalEndpoint:
      process.env.NEXT_PUBLIC_KRATOS_INTERNAL_ENDPOINT ??
      'http://kratos-public',
    homeUrl: process.env.NEXT_PUBLIC_HOME_URL ?? 'https://tadoku.app',
  },
  transpilePackages: ['ui'],
}

module.exports = nextConfig
