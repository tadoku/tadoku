/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  serverRuntimeConfig: {
    apiEndpoint:
      process.env.NEXT_SERVER_API_ENDPOINT ??
      process.env.NEXT_PUBLIC_API_ENDPOINT ??
      'https://tadoku.app/api/internal',
  },
  publicRuntimeConfig: {
    homeUrl: process.env.NEXT_PUBLIC_HOME_URL ?? 'https://tadoku.app',
    adminUrl: process.env.NEXT_PUBLIC_ADMIN_URL ?? 'https://admin.tadoku.app',
    apiEndpoint:
      process.env.NEXT_PUBLIC_API_ENDPOINT ?? 'https://tadoku.app/api/internal',
  },
  transpilePackages: ['ui', 'api-client'],
  async redirects() {
    return [
      {
        source: '/contests',
        destination: '/contests/official',
        permanent: true,
      },
    ]
  },
}

module.exports = nextConfig
