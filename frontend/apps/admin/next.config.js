/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  publicRuntimeConfig: {
    kratosPublicEndpoint:
      process.env.NEXT_PUBLIC_KRATOS_ENDPOINT ??
      'https://account.tadoku.app/kratos',
    kratosInternalEndpoint:
      process.env.NEXT_PUBLIC_KRATOS_INTERNAL_ENDPOINT ??
      'http://kratos-public',
    authUiUrl:
      process.env.NEXT_PUBLIC_AUTH_UI_URL ?? 'https://account.tadoku.app',
    homeUrl: process.env.NEXT_PUBLIC_HOME_URL ?? 'https://tadoku.app',
    apiEndpoint:
      process.env.NEXT_PUBLIC_API_ENDPOINT ?? 'https://tadoku.app/api/internal',
    cookieDomain: process.env.NEXT_PUBLIC_COOKIE_DOMAIN ?? '.tadoku.app',
  },
  transpilePackages: ['ui'],
}

module.exports = nextConfig
