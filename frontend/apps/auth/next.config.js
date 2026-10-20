/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  serverRuntimeConfig: {
    apiEndpoint: process.env.NEXT_SERVER_API_ENDPOINT ?? process.env.NEXT_PUBLIC_API_ENDPOINT ?? 'https://tadoku.app/api/internal',
  },
  publicRuntimeConfig: {
    // TODO: Figure out why this isn't getting passed to the client despite being opted-out from automatic static optimization
    kratosPublicEndpoint:
      process.env.NEXT_PUBLIC_KRATOS_ENDPOINT ??
      'https://account.tadoku.app/kratos',
    kratosInternalEndpoint:
      process.env.NEXT_PUBLIC_KRATOS_INTERNAL_ENDPOINT ??
      'http://kratos-public',
    homeUrl: process.env.NEXT_PUBLIC_HOME_URL ?? 'https://tadoku.app',
    authUiUrl: process.env.NEXT_PUBLIC_AUTH_UI_URL ?? 'https://account.tadoku.app',
    authV2UiUrl: process.env.NEXT_PUBLIC_AUTHV2_UI_URL,
    adminUrl: process.env.NEXT_PUBLIC_ADMIN_URL ?? 'https://admin.tadoku.app',
    apiEndpoint: process.env.NEXT_PUBLIC_API_ENDPOINT ?? 'https://tadoku.app/api/internal',
    cookieDomain: process.env.NEXT_PUBLIC_COOKIE_DOMAIN ?? '.tadoku.app',
    cookieSecure: process.env.NEXT_PUBLIC_COOKIE_SECURE !== 'false',
  },
  transpilePackages: ['ui', 'api-client'],
}

module.exports = nextConfig
